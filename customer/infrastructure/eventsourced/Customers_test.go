package eventsourced_test

import (
	"fmt"
	"go-iddd/customer/domain"
	"go-iddd/customer/domain/commands"
	"go-iddd/customer/domain/events"
	"go-iddd/customer/domain/values"
	"go-iddd/customer/infrastructure/eventsourced/test"
	"go-iddd/shared"
	"testing"

	"github.com/cockroachdb/errors"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCustomers_Register(t *testing.T) {
	Convey("Given a Repository", t, func() {
		diContainer := test.SetUpDIContainer()
		db := diContainer.GetPostgresDBConn()
		repo := diContainer.GetCustomerRepository()

		Convey("And given a new Customer", func() {
			id := values.GenerateCustomerID()
			recordedEvents := registerCustomerForCustomersTest(id)

			Convey("When the Customer is registered", func() {
				tx := test.BeginTx(db)
				session := repo.StartSession(tx)

				err := session.Register(id, recordedEvents)

				Convey("It should succeed", func() {
					So(err, ShouldBeNil)
					err = tx.Commit()
					So(err, ShouldBeNil)

					Convey("And when the same Customer is registered again", func() {
						customer := registerCustomerForCustomersTest(id)
						tx := test.BeginTx(db)
						session := repo.StartSession(tx)

						err = session.Register(id, customer)

						Convey("It should fail", func() {
							So(errors.Is(err, shared.ErrDuplicate), ShouldBeTrue)
						})
					})
				})
			})

			Convey("And given the session was already committed", func() {
				recordedEvents := registerCustomerForCustomersTest(id)
				tx := test.BeginTx(db)
				session := repo.StartSession(tx)
				err := tx.Commit()
				So(err, ShouldBeNil)

				Convey("When the Customer is registered", func() {
					err = session.Register(id, recordedEvents)

					Convey("It should fail", func() {
						So(errors.Is(err, shared.ErrTechnical), ShouldBeTrue)
					})
				})
			})

			cleanUpArtefactsForCustomers(id)
		})

		Convey("And given an existing Customer", func() {
			id := values.GenerateCustomerID()
			recordedEvents := registerCustomerForCustomersTest(id)
			tx := test.BeginTx(db)
			session := repo.StartSession(tx)
			err := session.Register(id, recordedEvents)
			So(err, ShouldBeNil)
			err = tx.Commit()
			So(err, ShouldBeNil)

			Convey("When the same Customer is registered again", func() {
				recordedEvents := registerCustomerForCustomersTest(id)
				tx := test.BeginTx(db)
				session := repo.StartSession(tx)

				err = session.Register(id, recordedEvents)

				Convey("It should fail", func() {
					So(errors.Is(err, shared.ErrDuplicate), ShouldBeTrue)
				})
			})

			cleanUpArtefactsForCustomers(id)
		})
	})
}

func TestCustomers_Of(t *testing.T) {
	Convey("Given an existing Customer", t, func() {
		id := values.GenerateCustomerID()
		eventStream := registerCustomerForCustomersTest(id)
		diContainer := test.SetUpDIContainer()
		db := diContainer.GetPostgresDBConn()
		repo := diContainer.GetCustomerRepository()
		// store := diContainer.GetPostgresEventStore()
		tx := test.BeginTx(db)
		session := repo.StartSession(tx)
		err := session.Register(id, eventStream)
		So(err, ShouldBeNil)
		err = tx.Commit()
		So(err, ShouldBeNil)

		Convey("When the Customer is retrieved", func() {
			session := repo.StartSession(tx)

			eventStream, err := session.EventStream(id)

			Convey("It should succeed", func() {
				So(err, ShouldBeNil)
				So(eventStream, ShouldHaveSameTypeAs, shared.DomainEvents{})
				So(eventStream, ShouldHaveLength, 1)
			})
		})

		Convey("And given the DB connection was closed", func() {
			tx := test.BeginTx(db)
			session := repo.StartSession(tx)

			err = db.Close()
			So(err, ShouldBeNil)

			Convey("When the Customer is retrieved", func() {
				eventStream, err := session.EventStream(id)

				Convey("It should fail", func() {
					So(errors.Is(err, shared.ErrTechnical), ShouldBeTrue)
					So(eventStream, ShouldHaveLength, 0)
				})
			})
		})

		cleanUpArtefactsForCustomers(id)
	})

	Convey("Given a not existing Customer", t, func() {
		id := values.GenerateCustomerID()
		diContainer := test.SetUpDIContainer()
		db := diContainer.GetPostgresDBConn()
		repo := diContainer.GetCustomerRepository()

		Convey("When the Customer is retrieved", func() {
			tx := test.BeginTx(db)
			session := repo.StartSession(tx)

			eventStream, err := session.EventStream(id)

			Convey("It should fail", func() {
				So(errors.Is(err, shared.ErrNotFound), ShouldBeTrue)
				So(eventStream, ShouldHaveLength, 0)
			})
		})
	})
}

func TestCustomers_Persist(t *testing.T) {
	Convey("Given a changed Customer", t, func() {
		id := values.GenerateCustomerID()
		recordedEvents := registerCustomerForCustomersTest(id)
		diContainer := test.SetUpDIContainer()
		db := diContainer.GetPostgresDBConn()
		repo := diContainer.GetCustomerRepository()
		tx := test.BeginTx(db)
		session := repo.StartSession(tx)
		err := session.Register(id, recordedEvents)
		So(err, ShouldBeNil)
		err = tx.Commit()
		So(err, ShouldBeNil)
		changeEmailAddress, err := commands.NewChangeEmailAddress(
			id.ID(),
			fmt.Sprintf("john+%s+changed@doe.com", id.ID()),
		)
		So(err, ShouldBeNil)

		recordedEvents = domain.ChangeEmailAddress(recordedEvents, changeEmailAddress)

		Convey("When the Customer is persisted", func() {
			tx := test.BeginTx(db)
			session := repo.StartSession(tx)

			err = session.Persist(id, recordedEvents)

			Convey("It should succeed", func() {
				So(err, ShouldBeNil)
				err = tx.Commit()
				So(err, ShouldBeNil)

				tx := test.BeginTx(db)
				session := repo.StartSession(tx)
				eventStream, err := session.EventStream(id)
				So(err, ShouldBeNil)
				So(eventStream, ShouldHaveSameTypeAs, shared.DomainEvents{})
				So(eventStream, ShouldHaveLength, 2)
				err = tx.Commit()
				So(err, ShouldBeNil)
			})
		})

		Convey("And given the session was already committed", func() {
			tx := test.BeginTx(db)
			session := repo.StartSession(tx)
			So(err, ShouldBeNil)

			err = tx.Commit()
			So(err, ShouldBeNil)

			Convey("When the Customer is persisted", func() {
				err = session.Persist(id, recordedEvents)

				Convey("It should fail", func() {
					So(errors.Is(err, shared.ErrTechnical), ShouldBeTrue)
				})
			})
		})

		cleanUpArtefactsForCustomers(id)
	})
}

/*** Test Helper Methods ***/

func registerCustomerForCustomersTest(id values.CustomerID) shared.DomainEvents {
	emailAddress := fmt.Sprintf("john+%s@doe.com", id.ID())
	givenName := "John"
	familyName := "Doe"
	register, err := commands.NewRegister(id.ID(), emailAddress, givenName, familyName)
	So(err, ShouldBeNil)

	recordedEvents := domain.RegisterCustomer(register)
	So(recordedEvents, ShouldHaveLength, 1)
	So(recordedEvents[0], ShouldHaveSameTypeAs, events.Registered{})

	return recordedEvents
}

func cleanUpArtefactsForCustomers(id values.CustomerID) {
	diContainer := test.SetUpDIContainer()
	store := diContainer.GetPostgresEventStore()

	streamID := shared.NewStreamID("customer" + "-" + id.ID())
	err := store.PurgeEventStream(streamID)
	So(err, ShouldBeNil)
}
