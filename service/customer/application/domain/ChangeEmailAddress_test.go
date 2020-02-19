package domain_test

import (
	"go-iddd/service/customer/application/domain"
	"go-iddd/service/customer/application/domain/commands"
	"go-iddd/service/customer/application/domain/events"
	"go-iddd/service/customer/application/domain/values"
	"go-iddd/service/lib/es"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestChangeEmailAddress(t *testing.T) {
	Convey("Prepare test artifacts", t, func() {
		customerID := values.GenerateCustomerID()
		emailAddress := values.RebuildEmailAddress("kevin@ball.com")
		confirmationHash := values.GenerateConfirmationHash(emailAddress.EmailAddress())
		personName := values.RebuildPersonName("Kevin", "Ball")
		changedEmailAddress := values.RebuildEmailAddress("latoya@ball.net")
		changedConfirmationHash := values.GenerateConfirmationHash(changedEmailAddress.EmailAddress())

		registered := events.ItWasRegistered(
			customerID,
			emailAddress,
			confirmationHash,
			personName,
			1,
		)

		emailAddressConfirmed := events.EmailAddressWasConfirmed(
			customerID,
			emailAddress,
			2,
		)

		changeEmailAddress, err := commands.NewChangeEmailAddress(
			customerID.ID(),
			changedEmailAddress.EmailAddress(),
		)
		So(err, ShouldBeNil)

		confirmEmailAddress, err := commands.NewConfirmEmailAddress(
			customerID.ID(),
			changedEmailAddress.EmailAddress(),
			changedConfirmationHash.Hash(),
		)
		So(err, ShouldBeNil)

		Convey("\nSCENARIO 1: Change a Customer's emailAddress", func() {
			Convey("Given CustomerRegistered", func() {
				eventStream := es.DomainEvents{registered}

				Convey("When ChangeEmailAddress", func() {
					recordedEvents := domain.ChangeEmailAddress(eventStream, changeEmailAddress)

					Convey("Then EmailAddressChanged", func() {
						ThenEmailAddressChanged(recordedEvents, changeEmailAddress, 2)
					})
				})
			})
		})

		Convey("\nSCENARIO 2: Try to change a Customer's emailAddress to the same value twice", func() {
			Convey("Given CustomerRegistered", func() {
				eventStream := es.DomainEvents{registered}

				Convey("and EmailAddressChanged", func() {
					emailAddressChanged := events.EmailAddressWasChanged(
						customerID,
						changedEmailAddress,
						changedConfirmationHash,
						2,
					)

					eventStream = append(eventStream, emailAddressChanged)

					Convey("When ChangeEmailAddress", func() {
						recordedEvents := domain.ChangeEmailAddress(eventStream, changeEmailAddress)

						Convey("Then no event", func() {
							So(recordedEvents, ShouldBeEmpty)
						})
					})
				})
			})
		})

		Convey("\nSCENARIO 3: Confirm a Customer's changed emailAddress, after the original emailAddress was confirmed", func() {
			Convey("Given CustomerRegistered", func() {
				eventStream := es.DomainEvents{registered}

				Convey("and EmailAddressConfirmed", func() {
					eventStream = append(eventStream, emailAddressConfirmed)

					Convey("and EmailAddressChanged", func() {
						emailAddressChanged := events.EmailAddressWasChanged(
							customerID,
							changedEmailAddress,
							changedConfirmationHash,
							3,
						)

						eventStream = append(eventStream, emailAddressChanged)

						Convey("When ConfirmEmailAddress", func() {
							recordedEvents := domain.ConfirmEmailAddress(eventStream, confirmEmailAddress)

							Convey("Then EmailAddressConfirmed", func() {
								ThenChangedEmailAddressConfirmed(recordedEvents, confirmEmailAddress, 4)
							})
						})
					})
				})
			})
		})
	})
}

func ThenEmailAddressChanged(
	recordedEvents es.DomainEvents,
	changeEmailAddress commands.ChangeEmailAddress,
	streamVersion uint,
) {

	So(recordedEvents, ShouldHaveLength, 1)
	emailAddressChanged, ok := recordedEvents[0].(events.EmailAddressChanged)
	So(ok, ShouldBeTrue)
	So(emailAddressChanged, ShouldNotBeNil)
	So(emailAddressChanged.CustomerID().Equals(changeEmailAddress.CustomerID()), ShouldBeTrue)
	So(emailAddressChanged.EmailAddress().Equals(changeEmailAddress.EmailAddress()), ShouldBeTrue)
	So(emailAddressChanged.ConfirmationHash().Equals(changeEmailAddress.ConfirmationHash()), ShouldBeTrue)
	So(emailAddressChanged.StreamVersion(), ShouldEqual, streamVersion)
}

func ThenChangedEmailAddressConfirmed(
	recordedEvents es.DomainEvents,
	confirmEmailAddress commands.ConfirmEmailAddress,
	streamVersion uint,
) {

	So(recordedEvents, ShouldHaveLength, 1)
	emailAddressConfirmed, ok := recordedEvents[0].(events.EmailAddressConfirmed)
	So(ok, ShouldBeTrue)
	So(emailAddressConfirmed.CustomerID().Equals(confirmEmailAddress.CustomerID()), ShouldBeTrue)
	So(emailAddressConfirmed.EmailAddress().Equals(confirmEmailAddress.EmailAddress()), ShouldBeTrue)
	So(emailAddressConfirmed.StreamVersion(), ShouldEqual, streamVersion)
}
