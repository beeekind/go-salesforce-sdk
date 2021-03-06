// Package apex contains helpers for executing anonymous Apex code
package apex

import (
	"bytes"
	"fmt"
	"net/http"
	"text/template"

	"github.com/beeekind/go-salesforce-sdk/requests"
)

const executeAnonymousEndpoint = "tooling/executeAnonymous"

var emailTemplate = template.Must(template.New("email.gohtml").
	Funcs(template.FuncMap{
		"First": func(elems []string) (string, error) { 
			if len(elems) == 0 { return "", nil }
			return elems[0], nil 
		 },
	}).
	Parse(`
		List<Map<String,String>> data = new List<Map<String,String>>();

		{{- range .}}
		data.add(new Map<String,String>{
			'toAddresses' => '{{.ToAddresses | First}}',
			'senderDisplayName'=> '{{.SenderDisplayName}}', 
			'subject' => '{{.Subject}}',
			'htmlBody' => '{{.HTMLBody}}'
		});
		{{- end}}

		List<Messaging.SingleEmailMessage> emails = new List<Messaging.SingleEmailMessage>();

		for (Map<String,String> email: data){
			Messaging.SingleEmailMessage message = new Messaging.SingleEmailMessage();
			
		    message.optOutPolicy = 'FILTER';
			message.toAddresses = new String[] { email.get('toAddresses') };
			message.senderDisplayName = email.get('senderDisplayName');
			message.subject = email.get('subject');
			message.htmlBody = email.get('htmlBody');

			emails.add(message);
		}

		Messaging.SendEmailResult[] results = Messaging.sendEmail(emails);`),
)

// SingleEmailMessage ...
// https://developer.salesforce.com/docs/atlas.en-us.api.meta/api/sforce_api_calls_sendemail.htm
type SingleEmailMessage struct {
	// Optional. An array of blind carbon copy (BCC) addresses or object IDs of the contacts, leads, and users you’re sending the email to.
	// This argument is allowed only when a template is not used.
	// The maximum size for this field is 4,000 bytes.
	// The maximum total of toAddresses, ccAddresses, and bccAddresses per email is 150.
	// All recipients in these three fields count against the limit for email sent using Apex or the API.
	//
	// You can specify opt-out email options with the optOutPolicy field only for those recipients who were added by their IDs.
	//
	// Email addresses are verified to ensure that they have the correct format and haven’t been marked as bounced.
	//
	// If the BCC COMPLIANCE option is set at the organization level, the user cannot add BCC addresses on standard messages.
	// The following error code is returned: BCC_NOT_ALLOWED_IF_BCC_COMPLIANCE_ENABLED.
	// All emails must have a recipient value in at least one of the following fields:
	//
	// * toAddresses
	// * ccAddresses
	// * bccAddresses
	// * targetObjectId
	BCCAddresses []string `json:"bccAddresses"`
	// BCCSender
	// Indicates whether the email sender receives a copy of the email that is sent. For a mass mail, the sender is only copied on the first email sent.
	BCCSender bool `json:"bccSender"`
	// CCAddresses
	//
	// Optional. An array of carbon copy (CC) addresses or object IDs of the contacts, leads, and users you’re sending the email to. This argument is allowed only when a template is not used.
	//
	// The maximum size for this field is 4,000 bytes. The maximum total of toAddresses, ccAddresses, and bccAddresses per email is 150. All recipients in these three fields count against the limit for email sent using Apex or the API.
	//
	// You can specify opt-out email options with the optOutPolicy field only for those recipients who were added by their IDs.
	//
	// Email addresses are verified to ensure that they have the correct format and haven’t been marked as bounced.
	//
	// All emails must have a recipient value in at least one of the following fields:
	// * toAddresses
	// * ccAddresses
	// * bccAddresses
	// * targetObjectId
	CCAddresses []string `json:"ccAddresses"`
	// Charset
	//
	// Optional. The character set for the email. If this value is null, the user's default value is used. Unavailable if specifying templateId because the template specifies the character set.
	Charset string `json:"charset"`
	// EmailPriority
	//
	// Optional. The priority of the email.
	// * Highest
	// * High
	// * Normal
	// * Low
	// * Lowest
	EmailPriority string `json:"emailPriority"`
	// EntityAttachments
	//
	// Optional. Array of IDs of Document, ContentVersion, or Attachment items to attach to the email. This field is available in API version 35.0 and later.
	EntityAttachments []string `json:"entityAttachments"`
	// FileAttachments
	// Optional. An array listing the file names of the binary and text files you want to attach to the email. You can attach multiple files as long as the total size of all attachments does not exceed 10 MB.
	//
	// This property has a broken documentation link: https://developer.salesforce.com/docs/#EmailFileAttachSection
	FileAttachments []*EmailFileAttachment `json:"fileAttachments"`
	// HTMLBody
	//
	// Optional. The HTML version of the email, specified by the sender. The value is encoded according to the specification associated with the organization.
	HTMLBody string `json:"htmlBody"`
	// InReplyTo
	//
	// Optional. The In-Reply-To field of the outgoing email. Identifies the emails to which this one is a reply (parent emails).
	// Contains the parent emails' Message-IDs. See RFC2822 - Internet Message Format.
	InReplyTo string `json:"inReplyTo"`
	// OptOutPlicy
	//
	// Optional. If you add contact, lead, or person account recipients by ID instead of email address, this field determines the behavior of the sendEmail() call.
	//
	// By default, the opt-out settings for recipients added by their email addresses aren’t checked and those recipients always receive the email.
	// Possible values of the SendEmailOptOutPolicy enumeration are:
	// * SEND - (default) The email is sent to all recipients. The recipients’ Email Opt Out setting is ignored. The setting Enforce email privacy settings is ignored.
	// * FILTER — No email is sent to recipients that have the Email Opt Out option set. Emails are sent to the other recipients. The setting Enforce email privacy settings is ignored.
	// * REJECT — If any of the recipients have the Email Opt Out option set, sendEmail() throws an error and no email is sent. The setting Enforce email privacy settings is respected, as are the selections in the data privacy record based on the Individual object. If any of the recipients have Don’t Market, Don’t Process, or Forget this Individual selected, sendEmail() throws an error and no email is sent.
	OptOutPolicy string `json:"optOutPolicy"`
	// OrgWideEmailAddressID
	//
	// Optional. The object ID of the OrgWideEmailAddress associated with the outgoing email. OrgWideEmailAddress.DisplayName cannot be set if the senderDisplayName field is already set.
	OrgWideEmailAddressID string `json:"orgWideEmailAddressId"`
	// PlainTextBody
	// Optional. The text version of the email, specified by the sender.
	PlainTextBody string `json:"plainTextBody"`
	// References
	// Optional. The References field of the outgoing email. Identifies an email thread. Contains the parent emails' Message-ID and References fields and possibly In-Reply-To fields. See RFC2822 - Internet Message Format.
	References string `json:"references"`
	// SaveAsActivity
	// Optional. The default value is true, meaning the email is saved as an activity. This argument only applies if the recipient list is based on targetObjectId or targetObjectIds. If HTML email tracking is enabled for the organization, you can track open rates.
	SaveAsActivity bool `json:"saveAsActivity"`
	// TargetObjectID
	// Optional. The object ID of the contact, lead, or user the email will be sent to.
	// The object ID you enter sets the context and ensures that merge fields in the template contain the correct data
	// All emails must have a recipient value in at least one of the following fields:
	// * toAddresses
	// * ccAddresses
	// * bccAddresses
	// * targetObjectId
	// SenderDisplayName
	// Optional. The name that appears on the From line of the email. This cannot be set if the object associated with a OrgWideEmailAddressId for a SingleEmailMessage has defined its DisplayName field.
	SenderDisplayName string `json:"senderDisplayName"`
	// Subject
	// Optional. The email subject line. If you are using an email template and attempt to override the subject line, an error message is returned.
	Subject        string `json:"subject"`
	TargetObjectID string `json:"targetObjectId"`
	// ToAddresses
	//
	// Optional. An array of email addresses or object IDs of the contacts, leads, or users you’re sending the email to. This argument is allowed only when a template is not used.
	//
	// The maximum size for this field is 4,000 bytes. The maximum total of toAddresses, ccAddresses, and bccAddresses per email is 150. All recipients in these three fields count against the limit for email sent using Apex or the API.
	//
	// You can specify opt-out email options with the optOutPolicy field only for those recipients who were added by their IDs.
	//
	// Email addresses are verified to ensure that they have the correct format and haven’t been marked as bounced.
	ToAddresses []string `json:"toAddresses"`
	// TreatBodiesAsTemplate
	// Optional. If set to true, the subject, plain text, and HTML text bodies of the email are treated as template data. The merge fields are resolved using the renderEmailTemplate() call. Default is false.
	// This field is available in API version 35.0 and later.
	TreatBodiesAsTemplate bool `json:"treatBodiesAsTemplate"`
	// TreatTargetObjectAsRecipient
	// Optional. If set to true, the targetObjectId (a contact, lead, or user) is the recipient of the email. If set to false, the targetObjectId is supplied as the WhoId field for template rendering but isn’t a recipient of the email. The default is true.
	//
	// This field is available in API version 35.0 and later. In prior versions, the targetObjectId is always a recipient of the email.
	TreatTargetObjectAsRecipient bool `json:"treatTargetObjectAsRecipient"`
	// WhatID
	// Optional. If you specify a contact for the targetObjectId field, you can specify a whatId as well. This field helps to further ensure that merge fields in the template contain the correct data.
	//
	// The value must be one of the following types:
	// * Account
	// * Asset
	// * Campaign
	// * Case
	// * Contract
	// * Opportunity
	// * Order
	// * Product
	// * Solution
	// * Custom
	WhatID string `json:"whatId"`
}

// EmailFileAttachment ...
//
// The following table contains properties that the EmailFileAttachment uses in the SingleEmailMessage object to specify attachments passed in as part of the request,
// as opposed to a Document passed in using the documentAttachments argument.
type EmailFileAttachment struct {
	// Body
	// The attachment itself
	Body string `json:"body"`
	// ContentType
	// Optional. The attachment's Content-Type
	ContentType string `json:"contentType"`
	// FileName
	// The name of the file to attach
	FileName string `json:"fileName"`
	// Inline
	// Optional. Specifies a Content-Disposition of inline (true) or attachment (false). In most cases, inline content is displayed to the user when the message is opened. Attachment content requires user action to be displayed.
	Inline bool `json:"inline"`
}

// SendEmailResponse ...
type SendEmailResponse struct {
	Column              int    `json:"column"`
	Compiled            bool   `json:"compiled"`
	CompileProblem      string `json:"compileProblem"`
	Line                int    `json:"line"`
	ExceptionMessage    string `json:"exceptionMessage"`
	ExceptionStacktrace string `json:"exceptionStackTrace"`
	Success             bool   `json:"success"`
}

// SendEmail ...
func SendEmail(req requests.Builder, emails ...*SingleEmailMessage) (response *SendEmailResponse, err error) {
	var buff bytes.Buffer
	if err := emailTemplate.Execute(&buff, emails); err != nil {
		return nil, err
	}

	_, err = req.
		URL(executeAnonymousEndpoint).
		Method(http.MethodGet).
		Param("anonymousBody", buff.String()).
		JSON(&response)

	if err != nil {
		return nil, err
	}

	if !response.Success {
		return response, fmt.Errorf("%s:%s:%s", response.CompileProblem, response.ExceptionMessage, response.ExceptionStacktrace)
	}

	return response, nil 
}
