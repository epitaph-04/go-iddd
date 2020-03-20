package customer

import (
	"github.com/AntonStoeckl/go-iddd/service/customer/domain/customer/commands"
	"github.com/AntonStoeckl/go-iddd/service/customer/domain/customer/events"
	"github.com/AntonStoeckl/go-iddd/service/customer/domain/customer/values"
	"github.com/AntonStoeckl/go-iddd/service/lib/es"
)

func ChangeName(eventStream es.DomainEvents, command commands.ChangeCustomerName) es.DomainEvents {
	var personName values.PersonName
	var currentStreamVersion uint

	for _, event := range eventStream {
		switch actualEvent := event.(type) {
		case events.CustomerRegistered:
			personName = actualEvent.PersonName()
		case events.CustomerNameChanged:
			personName = actualEvent.PersonName()
		}

		currentStreamVersion = event.StreamVersion()
	}

	if personName.Equals(command.PersonName()) {
		return nil
	}

	event := events.CustomerNameWasChanged(
		command.CustomerID(),
		command.PersonName(),
		currentStreamVersion+1,
	)

	return es.DomainEvents{event}
}