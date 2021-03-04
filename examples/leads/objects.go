package leads

// THIS PACKAGE IS AUTOGENERATED DO NOT EDIT

import "github.com/b3ntly/salesforce/types"

// Lead ...
//
// Represents a prospect or lead.
type Lead struct {
	// ConvertedDate ...
	//
	// Date on which this lead was converted.
	//
	// Properties:Filter, Group, Nillable, Sort
	ConvertedDate types.Date `json:"ConvertedDate"`
	// CompanyDunsNumber ...
	//
	// The Data Universal Numbering System (D-U-N-S) number, which is a unique, nine-digit number assigned to
	// every business location in the Dun & Bradstreet database that has a unique, separate, and distinct operation.
	// Industries and companies use D-U-N-S numbers as a global standard for business identification and tracking. Maximum
	// size is 9 characters. Note This field is only available to organizations that use Data.com Prospector or
	// Data.com Clean.
	//
	// Properties:Create, Filter, Group, Nillable, Sort, Update
	CompanyDunsNumber string `json:"CompanyDunsNumber"`
	// ProductInterest__c ...
	ProductInterest__c string `json:"ProductInterest__c"`
	// LastName ...
	//
	// Required. Last name of the lead up to 80 characters.
	//
	// Properties:Create, Filter, Group, Sort,
	// Update
	LastName string `json:"LastName"`
	// Country ...
	//
	// The lead’s country.
	//
	// Properties:Create, Filter, Group, Nillable, Sort, Update
	Country string `json:"Country"`
	// Address ...
	//
	// The compound form of the address. Read-only. For details on compound address fields, see Address Compound
	// Fields.
	//
	// Properties:Filter, Nillable
	Address types.Address `json:"Address"`
	// OwnerID ...
	//
	// ID of the lead’s owner.
	//
	// Properties:Create, Defaulted on create, Filter, Group, Sort, Update
	OwnerID string `json:"OwnerId"`
	// ConvertedContactID ...
	//
	// Object reference ID that points to the contact into which the lead converted.
	//
	//
	// Properties:Filter, Group, Nillable, Sort
	ConvertedContactID string `json:"ConvertedContactId"`
	// NumberofLocations__c ...
	NumberofLocations__c float64 `json:"NumberofLocations__c"`
	// State ...
	//
	// State for the address of the lead.
	//
	// Properties:Create, Filter, Group, Nillable, Sort, Update
	State string `json:"State"`
	// Status ...
	//
	// Status code for this converted lead. Status codes are defined in Status and represented in the API by the
	// LeadStatus object.
	//
	// Properties:Create, Defaulted on create, Filter, Group, Sort, Update
	Status string `json:"Status"`
	// LastModifiedDate ...
	LastModifiedDate types.Datetime `json:"LastModifiedDate"`
	// IsDeleted ...
	//
	// Indicates whether the object has been moved to the Recycle Bin (true) or not (false). Label is
	// Deleted.
	//
	// Properties:Defaulted on create, Filter
	IsDeleted bool `json:"IsDeleted"`
	// Phone ...
	//
	// The lead’s phone number.
	//
	// Properties:Create, Filter, Group, Nillable, Sort, Update
	Phone string `json:"Phone"`
	// LastReferencedDate ...
	//
	// The timestamp when the current user last accessed this record, a record related to this record, or a list
	// view.
	//
	// Properties:Filter, Nillable, Sort
	LastReferencedDate types.Datetime `json:"LastReferencedDate"`
	// FavoriteBalloon__c ...
	FavoriteBalloon__c string `json:"FavoriteBalloon__c"`
	// MobilePhone ...
	//
	// The lead’s mobile phone number.
	//
	// Properties:Create, Filter, Group, Nillable, Sort, Update
	MobilePhone string `json:"MobilePhone"`
	// Description ...
	//
	// The lead’s description.
	//
	// Properties:Create, Nillable, Update
	Description string `json:"Description"`
	// Industry ...
	//
	// Industry in which the lead works.
	//
	// Properties:Create, Filter, Group, Nillable, Sort, Update
	Industry string `json:"Industry"`
	// Jigsaw ...
	//
	// References the ID of a contact in Data.com. If a lead has a value in this field, it means that a contact was
	// imported as a lead from Data.com. If the contact (converted to a lead) was not imported from Data.com, the field value
	// is null. Maximum size is 20 characters. Available in API version 22.0 and later. Label is Data.com Key.
	// Important The Jigsawfield is exposed in the API to support troubleshooting for import errors and reimporting of
	// corrected data. Do not modify the value in the Jigsaw field.
	//
	// Properties:Create, Filter, Group, Nillable,
	// Sort, Update
	Jigsaw string `json:"Jigsaw"`
	// Street ...
	//
	// Street number and name for the address of the lead.
	//
	// Properties:Create, Filter, Group, Nillable,
	// Sort, Update
	Street string `json:"Street"`
	// Email ...
	//
	// The lead’s email address.
	//
	// Properties:Create, Filter, Group, idLookup, Nillable, Sort,
	// Update
	Email string `json:"Email"`
	// CurrentGenerators__c ...
	CurrentGenerators__c string `json:"CurrentGenerators__c"`
	// GeocodeAccuracy ...
	//
	// Accuracy level of the geocode for the address. For details on geolocation compound fields, see Compound
	// Field Considerations and Limitations.
	//
	// Properties:Create, Filter, Group, Retrieve, Query,
	// Restricted picklist, Nillable, Sort, Update
	GeocodeAccuracy string `json:"GeocodeAccuracy"`
	// JigsawContactID ...
	JigsawContactID string `json:"JigsawContactId"`
	// Primary__c ...
	Primary__c string `json:"Primary__c"`
	// PostalCode ...
	//
	// Postal code for the address of the lead. Label is Zip/Postal Code.
	//
	// Properties:Create, Filter,
	// Group, Nillable, Sort, Update
	PostalCode string `json:"PostalCode"`
	// NumberOfEmployees ...
	//
	// Number of employees at the lead’s company. Label is Employees.
	//
	// Properties:Create, Filter,
	// Group, Nillable, Sort, Update
	NumberOfEmployees int `json:"NumberOfEmployees"`
	// CreatedByID ...
	CreatedByID string `json:"CreatedById"`
	// MasterRecordID ...
	//
	// If this record was deleted as the result of a merge, this field contains the ID of the record that was kept. If
	// this record was deleted for any other reason, or has not been deleted, the value is null. Note When using Apex
	// triggers to determine which record was deleted in a merge event, this field’s value is the ID of the record that
	// remains in Trigger.old. In Trigger.new, the value is null.
	//
	// Properties:Filter, Group, Nillable, Sort
	MasterRecordID string `json:"MasterRecordId"`
	// Company ...
	//
	// Required. The lead’s company. Note If person account record types have been enabled, and if the value of
	// Company is null, the lead converts to a person account.
	//
	// Properties:Create, Filter, Group, Sort, Update
	Company string `json:"Company"`
	// IsConverted ...
	//
	// Indicates whether the lead has been converted (true) or not (false). Label is Converted.
	//
	//
	// Properties:Create, Defaulted on create, Filter, Group, Sort
	IsConverted bool `json:"IsConverted"`
	// EmailBouncedReason ...
	//
	// If bounce management is activated and an email sent to the lead bounced, the reason for the bounce.
	//
	//
	// Properties:Filter, Group, Nillable, Sort, Update
	EmailBouncedReason string `json:"EmailBouncedReason"`
	// FirstName ...
	//
	// The lead’s first name up to 40 characters.
	//
	// Properties:Create, Filter, Group, Nillable, Sort,
	// Update
	FirstName string `json:"FirstName"`
	// City ...
	//
	// City for the lead’s address.
	//
	// Properties:Create, Filter, Group, Nillable, Sort, Update
	City string `json:"City"`
	// PhotoURL ...
	//
	// Path to be combined with the URL of a Salesforce instance (Example:
	// https://yourInstance.salesforce.com/) to generate a URL to request the social network profile image associated with the lead. Generated URL
	// returns an HTTP redirect (code 302) to the social network profile image for the lead. Empty if Social Accounts and
	// Contacts isn't enabled or if Social Accounts and Contacts has been disabled for the requesting user.
	//
	//
	// Properties:Filter, Group, Nillable, Sort
	PhotoURL string `json:"PhotoUrl"`
	// ConvertedOpportunityID ...
	//
	// Object reference ID that points to the opportunity into which the lead has been converted.
	//
	//
	// Properties:Filter, Group, Nillable, Sort
	ConvertedOpportunityID string `json:"ConvertedOpportunityId"`
	// Salutation ...
	//
	// Salutation for the lead.
	//
	// Properties:Create, Filter, Group, Nillable, Sort, Update
	Salutation string `json:"Salutation"`
	// LeadSource ...
	//
	// The lead’s source.
	//
	// Properties:Create, Filter, Group, Nillable, Sort, Update
	LeadSource string `json:"LeadSource"`
	// ConvertedAccountID ...
	//
	// Object reference ID that points to the account into which the lead converted.
	//
	//
	// Properties:Filter, Group, Nillable, Sort
	ConvertedAccountID string `json:"ConvertedAccountId"`
	// SICCode__c ...
	SICCode__c string `json:"SICCode__c"`
	// ID ...
	ID string `json:"Id"`
	// Website ...
	//
	// Website for the lead.
	//
	// Properties:Create, Filter, Group, Nillable, Sort, Update
	Website string `json:"Website"`
	// SystemModstamp ...
	SystemModstamp types.Datetime `json:"SystemModstamp"`
	// EmailBouncedDate ...
	//
	// If bounce management is activated and an email sent to the lead bounced, the date and time of the
	// bounce.
	//
	// Properties:Filter, Nillable, Sort, Update
	EmailBouncedDate types.Datetime `json:"EmailBouncedDate"`
	// Actual_Do_Not_Call__c ...
	Actual_Do_Not_Call__c bool `json:"Actual_Do_Not_Call__c"`
	// Title ...
	//
	// Title for the lead, such as CFO or CEO.
	//
	// Properties:Create, Filter, Group, Nillable, Sort, Update
	Title string `json:"Title"`
	// Longitude ...
	//
	// Used with Latitude to specify the precise geolocation of an address. Acceptable values are numbers
	// between –180 and 180 up to 15 decimal places. For details on geolocation compound fields, see Compound Field
	// Considerations and Limitations.
	//
	// Properties:Create, Filter, Nillable, Sort, Update
	Longitude float64 `json:"Longitude"`
	// AnnualRevenue ...
	//
	// Annual revenue for the lead’s company.
	//
	// Properties:Create, Filter, Nillable, Sort, Update
	AnnualRevenue float64 `json:"AnnualRevenue"`
	// CreatedDate ...
	CreatedDate types.Datetime `json:"CreatedDate"`
	// IndividualID ...
	//
	// ID of the data privacy record associated with this lead. This field is available if you enabled Data
	// Protection and Privacy in Setup.
	//
	// Properties:Create, Filter, Group, Nillable, Sort, Update
	IndividualID string `json:"IndividualId"`
	// Name ...
	//
	// Concatenation of FirstName, MiddleName, LastName, and Suffix up to 203 characters, including
	// whitespaces.
	//
	// Properties:Filter, Group, Sort
	Name string `json:"Name"`
	// IsUnreadByOwner ...
	//
	// If true, lead has been assigned, but not yet viewed. See Unread Leads for more information. Label is Unread
	// By Owner.
	//
	// Properties:Create, Defaulted on create, Filter, Group, Sort, Update
	IsUnreadByOwner bool `json:"IsUnreadByOwner"`
	// CleanStatus ...
	//
	// Indicates the record’s clean status compared with Data.com. Values include: Matched, Different,
	// Acknowledged, NotFound, Inactive, Pending, SelectMatch, or Skipped.Several values for CleanStatus appear with
	// different labels on the lead record. Matched appears as In Sync Acknowledged appears as Reviewed Pending appears as
	// Not Compared
	//
	// Properties:Create, Filter, Group, Nillable, Restricted picklist, Sort, Update
	CleanStatus string `json:"CleanStatus"`
	// DandbCompanyID ...
	DandbCompanyID string `json:"DandbCompanyId"`
	// Latitude ...
	//
	// Used with Longitude to specify the precise geolocation of an address. Acceptable values are numbers
	// between –90 and 90 up to 15 decimal places. For details on geolocation compound fields, see Compound Field
	// Considerations and Limitations.
	//
	// Properties:Create, Filter, Nillable, Sort, Update
	Latitude float64 `json:"Latitude"`
	// Rating ...
	//
	// Rating of the lead.
	//
	// Properties:Create, Filter, Group, Nillable, Sort, Update
	Rating string `json:"Rating"`
	// Fax ...
	//
	// The lead’s fax number.
	//
	// Properties:Create, Filter, Group, Nillable, Sort, Update
	Fax string `json:"Fax"`
	// LastModifiedByID ...
	LastModifiedByID string `json:"LastModifiedById"`
	// LastActivityDate ...
	//
	// Value is the most recent of either: Due date of the most recent event logged against the record. Due date of
	// the most recently closed task associated with the record.
	//
	// Properties:Filter, Group, Nillable,
	// Sort
	LastActivityDate types.Date `json:"LastActivityDate"`
	// LastViewedDate ...
	//
	// The timestamp when the current user last viewed this record or list view. If this value is null, the user
	// might have only accessed this record or list view (LastReferencedDate) but not viewed it.
	//
	//
	// Properties:Filter, Nillable, Sort
	LastViewedDate types.Datetime `json:"LastViewedDate"`
}
