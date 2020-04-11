package serialization

import (
	"fmt"
	"testing"

	"github.com/AntonStoeckl/go-iddd/service/customer/domain/customer/events"
	"github.com/AntonStoeckl/go-iddd/service/customer/domain/customer/values"
	"github.com/AntonStoeckl/go-iddd/service/lib"
	"github.com/AntonStoeckl/go-iddd/service/lib/es"
	"github.com/cockroachdb/errors"
	. "github.com/smartystreets/goconvey/convey"
)

func TestMarshalAndUnmarshalCustomerEvents(t *testing.T) {
	customerID := values.GenerateCustomerID()
	emailAddress := values.RebuildEmailAddress("john@doe.com")
	newEmailAddress := values.RebuildEmailAddress("john.frank@doe.com")
	confirmationHash := values.GenerateConfirmationHash(emailAddress.String())
	personName := values.RebuildPersonName("John", "Doe")
	newPersonName := values.RebuildPersonName("John Frank", "Doe")
	failureReason := "wrong confirmation hash supplied"

	var myEvents es.DomainEvents
	streamVersion := uint(1)

	myEvents = append(
		myEvents,
		events.BuildCustomerRegistered(customerID, emailAddress, confirmationHash, personName, streamVersion),
	)

	streamVersion++

	myEvents = append(
		myEvents,
		events.BuildCustomerEmailAddressConfirmed(customerID, emailAddress, streamVersion),
	)

	streamVersion++

	myEvents = append(
		myEvents,
		events.BuildCustomerEmailAddressChanged(customerID, newEmailAddress, confirmationHash, emailAddress, streamVersion),
	)

	streamVersion++

	myEvents = append(
		myEvents,
		events.BuildCustomerNameChanged(customerID, newPersonName, streamVersion),
	)

	streamVersion++

	myEvents = append(
		myEvents,
		events.BuildCustomerDeleted(customerID, emailAddress, streamVersion),
	)

	for idx, event := range myEvents {
		originalEvent := event
		streamVersion = uint(idx + 1)
		eventName := originalEvent.Meta().EventName()

		Convey(fmt.Sprintf("When %s is marshaled and unmarshaled", eventName), t, func() {
			json, err := MarshalCustomerEvent(originalEvent)
			So(err, ShouldBeNil)

			unmarshaledEvent, err := UnmarshalCustomerEvent(originalEvent.Meta().EventName(), json, streamVersion)
			So(err, ShouldBeNil)

			Convey(fmt.Sprintf("Then the unmarshaled %s should resemble the original %s", eventName, eventName), func() {
				So(unmarshaledEvent, ShouldResemble, originalEvent)
			})
		})
	}

	// Special treatment for Failure events because the FailureReason()
	//  is a pointer to an error which does not resemble properly (ShouldResemble uses reflect.DeepEqual)

	Convey("When CustomerEmailAddressConfirmationFailed is marshaled and unmarshaled", t, func() {
		originalEvent := events.BuildCustomerEmailAddressConfirmationFailed(
			customerID, emailAddress, confirmationHash, errors.Mark(errors.New(failureReason), lib.ErrDomainConstraintsViolation), streamVersion,
		)

		oEventName := originalEvent.Meta().EventName()

		json, err := MarshalCustomerEvent(originalEvent)
		So(err, ShouldBeNil)

		unmarshaledEvent, err := UnmarshalCustomerEvent(originalEvent.Meta().EventName(), json, streamVersion)
		So(err, ShouldBeNil)

		uEventName := unmarshaledEvent.Meta().EventName()

		Convey(fmt.Sprintf("Then the unmarshaled %s should resemble the original %s", oEventName, uEventName), func() {
			unmarshaledEvent, ok := unmarshaledEvent.(events.CustomerEmailAddressConfirmationFailed)
			So(ok, ShouldBeTrue)
			So(unmarshaledEvent.CustomerID().Equals(originalEvent.CustomerID()), ShouldBeTrue)
			So(unmarshaledEvent.EmailAddress().Equals(originalEvent.EmailAddress()), ShouldBeTrue)
			So(unmarshaledEvent.ConfirmationHash().Equals(originalEvent.ConfirmationHash()), ShouldBeTrue)
			assertEventMetaResembles(originalEvent, unmarshaledEvent)
		})
	})
}

func assertEventMetaResembles(originalEvent es.DomainEvent, unmarshaledEvent es.DomainEvent) {
	So(unmarshaledEvent.Meta().EventName(), ShouldEqual, originalEvent.Meta().EventName())
	So(unmarshaledEvent.Meta().OccurredAt(), ShouldEqual, originalEvent.Meta().OccurredAt())
	So(unmarshaledEvent.Meta().StreamVersion(), ShouldEqual, originalEvent.Meta().StreamVersion())
	So(unmarshaledEvent.IsFailureEvent(), ShouldEqual, originalEvent.IsFailureEvent())
	So(unmarshaledEvent.FailureReason(), ShouldBeError)
	So(unmarshaledEvent.FailureReason().Error(), ShouldEqual, originalEvent.FailureReason().Error())
	So(errors.Is(originalEvent.FailureReason(), lib.ErrDomainConstraintsViolation), ShouldBeTrue)
	So(errors.Is(unmarshaledEvent.FailureReason(), lib.ErrDomainConstraintsViolation), ShouldBeTrue)
}

func TestMarshalCustomerEvent_WithUnknownEvent(t *testing.T) {
	Convey("When an unknown event is marshaled", t, func() {
		_, err := MarshalCustomerEvent(SomeEvent{})

		Convey("Then it should fail", func() {
			So(errors.Is(err, lib.ErrMarshalingFailed), ShouldBeTrue)
		})
	})
}

func TestUnmarshalCustomerEvent_WithUnknownEvent(t *testing.T) {
	Convey("When an unknown event is unmarshaled", t, func() {
		_, err := UnmarshalCustomerEvent("unknown", []byte{}, 1)

		Convey("Then it should fail", func() {
			So(errors.Is(err, lib.ErrUnmarshalingFailed), ShouldBeTrue)
		})
	})
}

/***** a mock event to test marshaling unknown event *****/

type SomeEvent struct{}

func (event SomeEvent) Meta() es.EventMeta {
	return es.RebuildEventMeta("SomeEvent", "never", 1)
}

func (event SomeEvent) IsFailureEvent() bool {
	return false
}

func (event SomeEvent) FailureReason() error {
	return nil
}