package commands

import (
	"go-iddd/customer/domain/values"
	"go-iddd/shared"
	"reflect"
	"strings"

	"github.com/cockroachdb/errors"
)

type Register struct {
	customerID   values.CustomerID
	emailAddress values.EmailAddress
	personName   values.PersonName
	isValid      bool
}

func NewRegister(
	customerID string,
	emailAddress string,
	givenName string,
	familyName string,
) (Register, error) {

	customerIDValue, err := values.BuildCustomerID(customerID)
	if err != nil {
		return Register{}, err
	}

	emailAddressValue, err := values.BuildEmailAddress(emailAddress)
	if err != nil {
		return Register{}, err
	}

	personNameValue, err := values.BuildPersonName(givenName, familyName)
	if err != nil {
		return Register{}, err
	}

	register := Register{
		customerID:   customerIDValue,
		emailAddress: emailAddressValue,
		personName:   personNameValue,
		isValid:      true,
	}

	return register, nil
}

func (register Register) CustomerID() values.CustomerID {
	return register.customerID
}

func (register Register) EmailAddress() values.EmailAddress {
	return register.emailAddress
}

func (register Register) PersonName() values.PersonName {
	return register.personName
}

func (register Register) AggregateID() shared.IdentifiesAggregates {
	return register.customerID
}

func (register Register) ShouldBeValid() error {
	if !register.isValid {
		err := errors.Newf("%s: is not valid", register.commandName())

		return errors.Mark(err, shared.ErrCommandIsInvalid)
	}

	return nil
}

func (register Register) commandName() string {
	commandType := reflect.TypeOf(register).String()
	commandTypeParts := strings.Split(commandType, ".")
	commandName := commandTypeParts[len(commandTypeParts)-1]

	return strings.Title(commandName)
}
