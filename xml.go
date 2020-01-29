package main

import (
	"time"
	"fmt"
)

const SOAP_TIME_FORMAT = "2006-01-02T15:04:05.00000Z"

func renderSubscribeXML(msgId string, callbackURL string, expiration time.Time) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>


<e:Envelope
	xmlns:e="http://www.w3.org/2003/05/soap-envelope"
	xmlns:wsnt="http://docs.oasis-open.org/wsn/b-2"
	xmlns:wsa5="http://www.w3.org/2005/08/addressing">
	<e:Header>
		<wsa5:MessageID>%s</wsa5:MessageID>
		<wsa5:Action e:mustUnderstand="true">http://docs.oasis-open.org/wsn/bw-2/NotificationProducer/SubscribeRequest</wsa5:Action>
	</e:Header>
	<e:Body>
		<wsnt:Subscribe>
			<wsnt:ConsumerReference>
				<wsa5:Address>%s</wsa5:Address>
			</wsnt:ConsumerReference>
			<wsnt:InitialTerminationTime>%s</wsnt:InitialTerminationTime>
		</wsnt:Subscribe>
	</e:Body>
</e:Envelope>
`, msgId, callbackURL, expiration.UTC().Format(SOAP_TIME_FORMAT))
}

func renderSubscriptionRenewXML(msgId string, expiration time.Time) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<e:Envelope
	xmlns:e="http://www.w3.org/2003/05/soap-envelope"
	xmlns:wsnt="http://docs.oasis-open.org/wsn/b-2"
	xmlns:wsa5="http://www.w3.org/2005/08/addressing">
	<e:Header>
		<wsa5:MessageID>%s</wsa5:MessageID>
		<wsa5:Action e:mustUnderstand="true">http://docs.oasis-open.org/wsn/bw-2/NotificationProducer/RenewRequest</wsa5:Action>
	</e:Header>
	<e:Body>
		<wsnt:Renew>
			<wsnt:TerminationTime>%s</wsnt:TerminationTime>
		</wsnt:Renew>
	</e:Body>
</e:Envelope>
`, msgId, expiration.UTC().Format(SOAP_TIME_FORMAT))
}

func renderUnsubscribeXML(msgId string) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<e:Envelope
	xmlns:e="http://www.w3.org/2003/05/soap-envelope"
	xmlns:wsnt="http://docs.oasis-open.org/wsn/b-2"
	xmlns:wsa5="http://www.w3.org/2005/08/addressing">
	<e:Header>
		<wsa5:MessageID>%s</wsa5:MessageID>
		<wsa5:Action e:mustUnderstand="true">http://docs.oasis-open.org/wsn/bw-2/NotificationProducer/UnsubscribeRequest</wsa5:Action>
	</e:Header>
	<e:Body>
		<wsnt:Unsubscribe></wsnt:Unsubscribe>
	</e:Body>
</e:Envelope>
`, msgId)
}