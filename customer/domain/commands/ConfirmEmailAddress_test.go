// Code generated by generate/main.go. DO NOT EDIT.

package commands_test

import (
	"go-iddd/customer/domain/commands"
	"go-iddd/customer/domain/values"
	"go-iddd/shared"
	"testing"

	"github.com/cockroachdb/errors"
	. "github.com/smartystreets/goconvey/convey"
)

func TestNewConfirmEmailAddress(t *testing.T) {
	Convey("Given valid input", t, func() {
		customerID := "64bcf656-da30-4f5a-b0b5-aead60965aa3"
		emailAddress := "john@doe.com"
		confirmationHash := "secret_hash"

		Convey("When a new ConfirmEmailAddress command is created", func() {
			confirmEmailAddress, err := commands.NewConfirmEmailAddress(customerID, emailAddress, confirmationHash)

			Convey("It should succeed", func() {
				So(err, ShouldBeNil)
				So(confirmEmailAddress, ShouldHaveSameTypeAs, (*commands.ConfirmEmailAddress)(nil))
			})
		})

		Convey("Given that customerID is invalid", func() {
			customerID = ""
			conveyNewConfirmEmailAddressWithInvalidInput(customerID, emailAddress, confirmationHash)
		})

		Convey("Given that emailAddress is invalid", func() {
			emailAddress = ""
			conveyNewConfirmEmailAddressWithInvalidInput(customerID, emailAddress, confirmationHash)
		})

		Convey("Given that confirmationHash is invalid", func() {
			confirmationHash = ""
			conveyNewConfirmEmailAddressWithInvalidInput(customerID, emailAddress, confirmationHash)
		})
	})
}

func conveyNewConfirmEmailAddressWithInvalidInput(
	customerID string,
	emailAddress string,
	confirmationHash string,
) {

	Convey("When a new ConfirmEmailAddress command is created", func() {
		confirmEmailAddress, err := commands.NewConfirmEmailAddress(customerID, emailAddress, confirmationHash)

		Convey("It should fail", func() {
			So(err, ShouldBeError)
			So(errors.Is(err, shared.ErrInputIsInvalid), ShouldBeTrue)
			So(confirmEmailAddress, ShouldBeNil)
		})
	})
}

func TestConfirmEmailAddressExposesExpectedValues(t *testing.T) {
	Convey("Given a ConfirmEmailAddress command", t, func() {
		customerID := "64bcf656-da30-4f5a-b0b5-aead60965aa3"
		emailAddress := "john@doe.com"
		confirmationHash := "secret_hash"

		customerIDValue, err := values.RebuildCustomerID(customerID)
		So(err, ShouldBeNil)
		emailAddressValue, err := values.NewEmailAddress(emailAddress)
		So(err, ShouldBeNil)
		confirmationHashValue, err := values.RebuildConfirmationHash(confirmationHash)
		So(err, ShouldBeNil)

		confirmEmailAddress, err := commands.NewConfirmEmailAddress(customerID, emailAddress, confirmationHash)
		So(err, ShouldBeNil)

		Convey("It should expose the expected values", func() {
			So(customerIDValue.Equals(confirmEmailAddress.CustomerID()), ShouldBeTrue)
			So(emailAddressValue.Equals(confirmEmailAddress.EmailAddress()), ShouldBeTrue)
			So(confirmationHashValue.Equals(confirmEmailAddress.ConfirmationHash()), ShouldBeTrue)
			So(confirmEmailAddress.CommandName(), ShouldEqual, "ConfirmEmailAddress")
			So(customerIDValue.Equals(confirmEmailAddress.AggregateID()), ShouldBeTrue)
		})
	})
}
