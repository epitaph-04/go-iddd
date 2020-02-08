package acceptance_test

import (
	"fmt"
	"go-iddd/service/customer/application/domain/commands"
	"go-iddd/service/customer/application/domain/events"
	"go-iddd/service/customer/application/domain/values"
	"go-iddd/service/customer/infrastructure"
	"go-iddd/service/customer/infrastructure/secondary/forstoringcustomerevents/eventstore"
	"go-iddd/service/lib"
	"testing"

	"github.com/cockroachdb/errors"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_CustomerLifecycle(t *testing.T) {
	Convey("Scenarios", t, func() {
		diContainer, err := infrastructure.SetUpDIContainer()
		So(err, ShouldBeNil)

		commandHandler := diContainer.GetCustomerCommandHandler()
		customerEventStore := diContainer.GetCustomerEventStore()

		emailAddress := "jane@doe.com"
		givenName := "Jane"
		familyName := "Doe"

		register, err := commands.NewRegister(
			emailAddress,
			givenName,
			familyName,
		)
		So(err, ShouldBeNil)

		Convey("SCENARIO: A prospective Customer registers and confirms her email address", func() {
			Convey(fmt.Sprintf("When I register as %s %s with %s", givenName, familyName, emailAddress), func() {
				err := commandHandler.Register(register)
				So(err, ShouldBeNil)

				Convey("and I confirm my email address with a valid confirmation hash", func() {
					confirmEmailAddress, err := commands.NewConfirmEmailAddress(
						register.CustomerID().ID(),
						register.EmailAddress().EmailAddress(),
						register.ConfirmationHash().Hash(),
					)
					So(err, ShouldBeNil)

					err = commandHandler.ConfirmEmailAddress(confirmEmailAddress)
					So(err, ShouldBeNil)

					Convey("Then my email address should be confirmed", func() {
						MyEmailAddressShouldBeConfirmed(register.CustomerID(), customerEventStore)

						Convey("When I try to confirm it again", func() {
							err = commandHandler.ConfirmEmailAddress(confirmEmailAddress)

							Convey("Then it should be ignored", func() {
								So(err, ShouldBeNil)
								MyEmailAddressShouldBeConfirmed(register.CustomerID(), customerEventStore)
							})
						})
					})
				})
			})
		})

		Convey("SCENARIO: A Customer fails to confirm her email address", func() {
			Convey("Given I registered as a Customer", func() {
				err := commandHandler.Register(register)
				So(err, ShouldBeNil)

				Convey("When I try to confirm my email address with an invalid confirmation hash", func() {
					confirmEmailAddress, err := commands.NewConfirmEmailAddress(
						register.CustomerID().ID(),
						register.EmailAddress().EmailAddress(),
						values.GenerateConfirmationHash(register.EmailAddress().EmailAddress()).Hash(),
					)
					So(err, ShouldBeNil)

					err = commandHandler.ConfirmEmailAddress(confirmEmailAddress)

					Convey("Then it should fail", func() {
						So(err, ShouldBeError)
						So(errors.Is(err, lib.ErrDomainConstraintsViolation), ShouldBeTrue)

						Convey("and my email address should be unconfirmed", func() {
							MyEmailAddressShouldNotBeConfirmed(register.CustomerID(), customerEventStore)
						})
					})
				})
			})
		})

		Convey("SCENARIO: A Customer changes her confirmed email address", func() {
			Convey("Given I registered as a Customer", func() {
				err := commandHandler.Register(register)
				So(err, ShouldBeNil)

				Convey("and I confirmed my email address", func() {
					confirmEmailAddress, err := commands.NewConfirmEmailAddress(
						register.CustomerID().ID(),
						register.EmailAddress().EmailAddress(),
						register.ConfirmationHash().Hash(),
					)
					So(err, ShouldBeNil)

					err = commandHandler.ConfirmEmailAddress(confirmEmailAddress)
					So(err, ShouldBeNil)

					Convey("When I change my email address", func() {
						changeEmailAddress, err := commands.NewChangeEmailAddress(
							register.CustomerID().ID(),
							"john@doe.com",
						)
						So(err, ShouldBeNil)

						err = commandHandler.ChangeEmailAddress(changeEmailAddress)
						So(err, ShouldBeNil)

						Convey("Then my changed email address should not be confirmed", func() {
							MyEmailChangedAddressShouldNotBeConfirmed(register.CustomerID(), customerEventStore)

							Convey("When I try to change it again", func() {
								changeEmailAddress, err := commands.NewChangeEmailAddress(
									register.CustomerID().ID(),
									"john@doe.com",
								)
								So(err, ShouldBeNil)

								err = commandHandler.ChangeEmailAddress(changeEmailAddress)
								Convey("Then it should be ignored", func() {
									So(err, ShouldBeNil)
									MyEmailChangedAddressShouldNotBeConfirmed(register.CustomerID(), customerEventStore)
								})
							})
						})
					})
				})
			})
		})

		Convey("SCENARIO: A Customer confirms a changed email address", func() {
			Convey("Given I registered as a Customer", func() {
				err := commandHandler.Register(register)
				So(err, ShouldBeNil)

				Convey("and I confirmed my email address", func() {
					confirmEmailAddress, err := commands.NewConfirmEmailAddress(
						register.CustomerID().ID(),
						register.EmailAddress().EmailAddress(),
						register.ConfirmationHash().Hash(),
					)
					So(err, ShouldBeNil)

					err = commandHandler.ConfirmEmailAddress(confirmEmailAddress)
					So(err, ShouldBeNil)

					Convey("and I changed my email address", func() {
						changeEmailAddress, err := commands.NewChangeEmailAddress(
							register.CustomerID().ID(),
							"john@doe.com",
						)
						So(err, ShouldBeNil)

						err = commandHandler.ChangeEmailAddress(changeEmailAddress)
						So(err, ShouldBeNil)

						Convey("When I confirm my changed email address", func() {
							confirmEmailAddress, err := commands.NewConfirmEmailAddress(
								changeEmailAddress.CustomerID().ID(),
								changeEmailAddress.EmailAddress().EmailAddress(),
								changeEmailAddress.ConfirmationHash().Hash(),
							)
							So(err, ShouldBeNil)

							err = commandHandler.ConfirmEmailAddress(confirmEmailAddress)
							So(err, ShouldBeNil)

							Convey("Then my changed email address should be confirmed", func() {
								MyEmailChangedAddressShouldBeConfirmed(register.CustomerID(), customerEventStore)
							})
						})
					})
				})
			})

		})

		Reset(func() {
			err := customerEventStore.Delete(register.CustomerID())
			So(err, ShouldBeNil)
		})
	})
}

func MyEmailAddressShouldBeConfirmed(customerID values.CustomerID, customerEventStore *eventstore.CustomerEventStore) {
	eventStream, err := customerEventStore.EventStreamFor(customerID)
	So(err, ShouldBeNil)
	So(eventStream, ShouldHaveLength, 2)
	So(eventStream[0], ShouldHaveSameTypeAs, events.Registered{})
	So(eventStream[1], ShouldHaveSameTypeAs, events.EmailAddressConfirmed{})
}

func MyEmailAddressShouldNotBeConfirmed(customerID values.CustomerID, customerEventStore *eventstore.CustomerEventStore) {
	eventStream, err := customerEventStore.EventStreamFor(customerID)
	So(err, ShouldBeNil)
	So(eventStream, ShouldHaveLength, 2)
	So(eventStream[0], ShouldHaveSameTypeAs, events.Registered{})
	So(eventStream[1], ShouldHaveSameTypeAs, events.EmailAddressConfirmationFailed{})
}

func MyEmailChangedAddressShouldBeConfirmed(customerID values.CustomerID, customerEventStore *eventstore.CustomerEventStore) {
	eventStream, err := customerEventStore.EventStreamFor(customerID)
	So(err, ShouldBeNil)
	So(eventStream, ShouldHaveLength, 4)
	So(eventStream[0], ShouldHaveSameTypeAs, events.Registered{})
	So(eventStream[1], ShouldHaveSameTypeAs, events.EmailAddressConfirmed{})
	So(eventStream[2], ShouldHaveSameTypeAs, events.EmailAddressChanged{})
	So(eventStream[3], ShouldHaveSameTypeAs, events.EmailAddressConfirmed{})
}

func MyEmailChangedAddressShouldNotBeConfirmed(customerID values.CustomerID, customerEventStore *eventstore.CustomerEventStore) {
	eventStream, err := customerEventStore.EventStreamFor(customerID)
	So(err, ShouldBeNil)
	So(eventStream, ShouldHaveLength, 3)
	So(eventStream[0], ShouldHaveSameTypeAs, events.Registered{})
	So(eventStream[1], ShouldHaveSameTypeAs, events.EmailAddressConfirmed{})
	So(eventStream[2], ShouldHaveSameTypeAs, events.EmailAddressChanged{})
}
