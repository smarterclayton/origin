package storage

// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Code generated by Microsoft (R) AutoRest Code Generator.
// Changes may cause incorrect behavior and will be lost if the code is regenerated.

import (
	"encoding/json"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/date"
	"net/http"
)

// AccessTier enumerates the values for access tier.
type AccessTier string

const (
	// Cool ...
	Cool AccessTier = "Cool"
	// Hot ...
	Hot AccessTier = "Hot"
)

// AccountStatus enumerates the values for account status.
type AccountStatus string

const (
	// Available ...
	Available AccountStatus = "available"
	// Unavailable ...
	Unavailable AccountStatus = "unavailable"
)

// HTTPProtocol enumerates the values for http protocol.
type HTTPProtocol string

const (
	// HTTPS ...
	HTTPS HTTPProtocol = "https"
	// Httpshttp ...
	Httpshttp HTTPProtocol = "https,http"
)

// KeyPermission enumerates the values for key permission.
type KeyPermission string

const (
	// Full ...
	Full KeyPermission = "Full"
	// Read ...
	Read KeyPermission = "Read"
)

// Kind enumerates the values for kind.
type Kind string

const (
	// BlobStorage ...
	BlobStorage Kind = "BlobStorage"
	// Storage ...
	Storage Kind = "Storage"
)

// Permissions enumerates the values for permissions.
type Permissions string

const (
	// A ...
	A Permissions = "a"
	// C ...
	C Permissions = "c"
	// D ...
	D Permissions = "d"
	// L ...
	L Permissions = "l"
	// P ...
	P Permissions = "p"
	// R ...
	R Permissions = "r"
	// U ...
	U Permissions = "u"
	// W ...
	W Permissions = "w"
)

// Permissions1 enumerates the values for permissions 1.
type Permissions1 string

const (
	// Permissions1A ...
	Permissions1A Permissions1 = "a"
	// Permissions1C ...
	Permissions1C Permissions1 = "c"
	// Permissions1D ...
	Permissions1D Permissions1 = "d"
	// Permissions1L ...
	Permissions1L Permissions1 = "l"
	// Permissions1P ...
	Permissions1P Permissions1 = "p"
	// Permissions1R ...
	Permissions1R Permissions1 = "r"
	// Permissions1U ...
	Permissions1U Permissions1 = "u"
	// Permissions1W ...
	Permissions1W Permissions1 = "w"
)

// ProvisioningState enumerates the values for provisioning state.
type ProvisioningState string

const (
	// Creating ...
	Creating ProvisioningState = "Creating"
	// ResolvingDNS ...
	ResolvingDNS ProvisioningState = "ResolvingDNS"
	// Succeeded ...
	Succeeded ProvisioningState = "Succeeded"
)

// Reason enumerates the values for reason.
type Reason string

const (
	// AccountNameInvalid ...
	AccountNameInvalid Reason = "AccountNameInvalid"
	// AlreadyExists ...
	AlreadyExists Reason = "AlreadyExists"
)

// ResourceEnum enumerates the values for resource enum.
type ResourceEnum string

const (
	// ResourceEnumB ...
	ResourceEnumB ResourceEnum = "b"
	// ResourceEnumC ...
	ResourceEnumC ResourceEnum = "c"
	// ResourceEnumF ...
	ResourceEnumF ResourceEnum = "f"
	// ResourceEnumS ...
	ResourceEnumS ResourceEnum = "s"
)

// ResourceTypes enumerates the values for resource types.
type ResourceTypes string

const (
	// ResourceTypesC ...
	ResourceTypesC ResourceTypes = "c"
	// ResourceTypesO ...
	ResourceTypesO ResourceTypes = "o"
	// ResourceTypesS ...
	ResourceTypesS ResourceTypes = "s"
)

// Services enumerates the values for services.
type Services string

const (
	// B ...
	B Services = "b"
	// F ...
	F Services = "f"
	// Q ...
	Q Services = "q"
	// T ...
	T Services = "t"
)

// SkuName enumerates the values for sku name.
type SkuName string

const (
	// PremiumLRS ...
	PremiumLRS SkuName = "Premium_LRS"
	// StandardGRS ...
	StandardGRS SkuName = "Standard_GRS"
	// StandardLRS ...
	StandardLRS SkuName = "Standard_LRS"
	// StandardRAGRS ...
	StandardRAGRS SkuName = "Standard_RAGRS"
	// StandardZRS ...
	StandardZRS SkuName = "Standard_ZRS"
)

// SkuTier enumerates the values for sku tier.
type SkuTier string

const (
	// Premium ...
	Premium SkuTier = "Premium"
	// Standard ...
	Standard SkuTier = "Standard"
)

// UsageUnit enumerates the values for usage unit.
type UsageUnit string

const (
	// Bytes ...
	Bytes UsageUnit = "Bytes"
	// BytesPerSecond ...
	BytesPerSecond UsageUnit = "BytesPerSecond"
	// Count ...
	Count UsageUnit = "Count"
	// CountsPerSecond ...
	CountsPerSecond UsageUnit = "CountsPerSecond"
	// Percent ...
	Percent UsageUnit = "Percent"
	// Seconds ...
	Seconds UsageUnit = "Seconds"
)

// Account the storage account.
type Account struct {
	autorest.Response `json:"-"`
	// ID - Resource Id
	ID *string `json:"id,omitempty"`
	// Name - Resource name
	Name *string `json:"name,omitempty"`
	// Type - Resource type
	Type *string `json:"type,omitempty"`
	// Location - Resource location
	Location *string `json:"location,omitempty"`
	// Tags - Tags assigned to a resource; can be used for viewing and grouping a resource (across resource groups).
	Tags *map[string]*string `json:"tags,omitempty"`
	// Sku - Gets the SKU.
	Sku *Sku `json:"sku,omitempty"`
	// Kind - Gets the Kind. Possible values include: 'Storage', 'BlobStorage'
	Kind               Kind `json:"kind,omitempty"`
	*AccountProperties `json:"properties,omitempty"`
}

// UnmarshalJSON is the custom unmarshaler for Account struct.
func (a *Account) UnmarshalJSON(body []byte) error {
	var m map[string]*json.RawMessage
	err := json.Unmarshal(body, &m)
	if err != nil {
		return err
	}
	var v *json.RawMessage

	v = m["sku"]
	if v != nil {
		var sku Sku
		err = json.Unmarshal(*m["sku"], &sku)
		if err != nil {
			return err
		}
		a.Sku = &sku
	}

	v = m["kind"]
	if v != nil {
		var kind Kind
		err = json.Unmarshal(*m["kind"], &kind)
		if err != nil {
			return err
		}
		a.Kind = kind
	}

	v = m["properties"]
	if v != nil {
		var properties AccountProperties
		err = json.Unmarshal(*m["properties"], &properties)
		if err != nil {
			return err
		}
		a.AccountProperties = &properties
	}

	v = m["id"]
	if v != nil {
		var ID string
		err = json.Unmarshal(*m["id"], &ID)
		if err != nil {
			return err
		}
		a.ID = &ID
	}

	v = m["name"]
	if v != nil {
		var name string
		err = json.Unmarshal(*m["name"], &name)
		if err != nil {
			return err
		}
		a.Name = &name
	}

	v = m["type"]
	if v != nil {
		var typeVar string
		err = json.Unmarshal(*m["type"], &typeVar)
		if err != nil {
			return err
		}
		a.Type = &typeVar
	}

	v = m["location"]
	if v != nil {
		var location string
		err = json.Unmarshal(*m["location"], &location)
		if err != nil {
			return err
		}
		a.Location = &location
	}

	v = m["tags"]
	if v != nil {
		var tags map[string]*string
		err = json.Unmarshal(*m["tags"], &tags)
		if err != nil {
			return err
		}
		a.Tags = &tags
	}

	return nil
}

// AccountCheckNameAvailabilityParameters the parameters used to check the availabity of the storage account name.
type AccountCheckNameAvailabilityParameters struct {
	Name *string `json:"name,omitempty"`
	Type *string `json:"type,omitempty"`
}

// AccountCreateParameters the parameters used when creating a storage account.
type AccountCreateParameters struct {
	// Sku - Required. Gets or sets the sku name.
	Sku *Sku `json:"sku,omitempty"`
	// Kind - Required. Indicates the type of storage account. Possible values include: 'Storage', 'BlobStorage'
	Kind Kind `json:"kind,omitempty"`
	// Location - Required. Gets or sets the location of the resource. This will be one of the supported and registered Azure Geo Regions (e.g. West US, East US, Southeast Asia, etc.). The geo region of a resource cannot be changed once it is created, but if an identical geo region is specified on update, the request will succeed.
	Location *string `json:"location,omitempty"`
	// Tags - Gets or sets a list of key value pairs that describe the resource. These tags can be used for viewing and grouping this resource (across resource groups). A maximum of 15 tags can be provided for a resource. Each tag must have a key with a length no greater than 128 characters and a value with a length no greater than 256 characters.
	Tags                               *map[string]*string `json:"tags,omitempty"`
	*AccountPropertiesCreateParameters `json:"properties,omitempty"`
}

// UnmarshalJSON is the custom unmarshaler for AccountCreateParameters struct.
func (acp *AccountCreateParameters) UnmarshalJSON(body []byte) error {
	var m map[string]*json.RawMessage
	err := json.Unmarshal(body, &m)
	if err != nil {
		return err
	}
	var v *json.RawMessage

	v = m["sku"]
	if v != nil {
		var sku Sku
		err = json.Unmarshal(*m["sku"], &sku)
		if err != nil {
			return err
		}
		acp.Sku = &sku
	}

	v = m["kind"]
	if v != nil {
		var kind Kind
		err = json.Unmarshal(*m["kind"], &kind)
		if err != nil {
			return err
		}
		acp.Kind = kind
	}

	v = m["location"]
	if v != nil {
		var location string
		err = json.Unmarshal(*m["location"], &location)
		if err != nil {
			return err
		}
		acp.Location = &location
	}

	v = m["tags"]
	if v != nil {
		var tags map[string]*string
		err = json.Unmarshal(*m["tags"], &tags)
		if err != nil {
			return err
		}
		acp.Tags = &tags
	}

	v = m["properties"]
	if v != nil {
		var properties AccountPropertiesCreateParameters
		err = json.Unmarshal(*m["properties"], &properties)
		if err != nil {
			return err
		}
		acp.AccountPropertiesCreateParameters = &properties
	}

	return nil
}

// AccountKey an access key for the storage account.
type AccountKey struct {
	// KeyName - Name of the key.
	KeyName *string `json:"keyName,omitempty"`
	// Value - Base 64-encoded value of the key.
	Value *string `json:"value,omitempty"`
	// Permissions - Permissions for the key -- read-only or full permissions. Possible values include: 'Read', 'Full'
	Permissions KeyPermission `json:"permissions,omitempty"`
}

// AccountListKeysResult the response from the ListKeys operation.
type AccountListKeysResult struct {
	autorest.Response `json:"-"`
	// Keys - Gets the list of storage account keys and their properties for the specified storage account.
	Keys *[]AccountKey `json:"keys,omitempty"`
}

// AccountListResult the response from the List Storage Accounts operation.
type AccountListResult struct {
	autorest.Response `json:"-"`
	// Value - Gets the list of storage accounts and their properties.
	Value *[]Account `json:"value,omitempty"`
}

// AccountProperties properties of the storage account.
type AccountProperties struct {
	// ProvisioningState - Gets the status of the storage account at the time the operation was called. Possible values include: 'Creating', 'ResolvingDNS', 'Succeeded'
	ProvisioningState ProvisioningState `json:"provisioningState,omitempty"`
	// PrimaryEndpoints - Gets the URLs that are used to perform a retrieval of a public blob, queue, or table object. Note that Standard_ZRS and Premium_LRS accounts only return the blob endpoint.
	PrimaryEndpoints *Endpoints `json:"primaryEndpoints,omitempty"`
	// PrimaryLocation - Gets the location of the primary data center for the storage account.
	PrimaryLocation *string `json:"primaryLocation,omitempty"`
	// StatusOfPrimary - Gets the status indicating whether the primary location of the storage account is available or unavailable. Possible values include: 'Available', 'Unavailable'
	StatusOfPrimary AccountStatus `json:"statusOfPrimary,omitempty"`
	// LastGeoFailoverTime - Gets the timestamp of the most recent instance of a failover to the secondary location. Only the most recent timestamp is retained. This element is not returned if there has never been a failover instance. Only available if the accountType is Standard_GRS or Standard_RAGRS.
	LastGeoFailoverTime *date.Time `json:"lastGeoFailoverTime,omitempty"`
	// SecondaryLocation - Gets the location of the geo-replicated secondary for the storage account. Only available if the accountType is Standard_GRS or Standard_RAGRS.
	SecondaryLocation *string `json:"secondaryLocation,omitempty"`
	// StatusOfSecondary - Gets the status indicating whether the secondary location of the storage account is available or unavailable. Only available if the SKU name is Standard_GRS or Standard_RAGRS. Possible values include: 'Available', 'Unavailable'
	StatusOfSecondary AccountStatus `json:"statusOfSecondary,omitempty"`
	// CreationTime - Gets the creation date and time of the storage account in UTC.
	CreationTime *date.Time `json:"creationTime,omitempty"`
	// CustomDomain - Gets the custom domain the user assigned to this storage account.
	CustomDomain *CustomDomain `json:"customDomain,omitempty"`
	// SecondaryEndpoints - Gets the URLs that are used to perform a retrieval of a public blob, queue, or table object from the secondary location of the storage account. Only available if the SKU name is Standard_RAGRS.
	SecondaryEndpoints *Endpoints `json:"secondaryEndpoints,omitempty"`
	// Encryption - Gets the encryption settings on the account. If unspecified, the account is unencrypted.
	Encryption *Encryption `json:"encryption,omitempty"`
	// AccessTier - Required for storage accounts where kind = BlobStorage. The access tier used for billing. Possible values include: 'Hot', 'Cool'
	AccessTier AccessTier `json:"accessTier,omitempty"`
	// EnableHTTPSTrafficOnly - Allows https traffic only to storage service if sets to true.
	EnableHTTPSTrafficOnly *bool `json:"supportsHttpsTrafficOnly,omitempty"`
}

// AccountPropertiesCreateParameters the parameters used to create the storage account.
type AccountPropertiesCreateParameters struct {
	// CustomDomain - User domain assigned to the storage account. Name is the CNAME source. Only one custom domain is supported per storage account at this time. To clear the existing custom domain, use an empty string for the custom domain name property.
	CustomDomain *CustomDomain `json:"customDomain,omitempty"`
	// Encryption - Provides the encryption settings on the account. If left unspecified the account encryption settings will remain the same. The default setting is unencrypted.
	Encryption *Encryption `json:"encryption,omitempty"`
	// AccessTier - Required for storage accounts where kind = BlobStorage. The access tier used for billing. Possible values include: 'Hot', 'Cool'
	AccessTier AccessTier `json:"accessTier,omitempty"`
	// EnableHTTPSTrafficOnly - Allows https traffic only to storage service if sets to true.
	EnableHTTPSTrafficOnly *bool `json:"supportsHttpsTrafficOnly,omitempty"`
}

// AccountPropertiesUpdateParameters the parameters used when updating a storage account.
type AccountPropertiesUpdateParameters struct {
	// CustomDomain - Custom domain assigned to the storage account by the user. Name is the CNAME source. Only one custom domain is supported per storage account at this time. To clear the existing custom domain, use an empty string for the custom domain name property.
	CustomDomain *CustomDomain `json:"customDomain,omitempty"`
	// Encryption - Provides the encryption settings on the account. The default setting is unencrypted.
	Encryption *Encryption `json:"encryption,omitempty"`
	// AccessTier - Required for storage accounts where kind = BlobStorage. The access tier used for billing. Possible values include: 'Hot', 'Cool'
	AccessTier AccessTier `json:"accessTier,omitempty"`
	// EnableHTTPSTrafficOnly - Allows https traffic only to storage service if sets to true.
	EnableHTTPSTrafficOnly *bool `json:"supportsHttpsTrafficOnly,omitempty"`
}

// AccountRegenerateKeyParameters the parameters used to regenerate the storage account key.
type AccountRegenerateKeyParameters struct {
	KeyName *string `json:"keyName,omitempty"`
}

// AccountSasParameters the parameters to list SAS credentials of a storage account.
type AccountSasParameters struct {
	// Services - The signed services accessible with the account SAS. Possible values include: Blob (b), Queue (q), Table (t), File (f). Possible values include: 'B', 'Q', 'T', 'F'
	Services Services `json:"signedServices,omitempty"`
	// ResourceTypes - The signed resource types that are accessible with the account SAS. Service (s): Access to service-level APIs; Container (c): Access to container-level APIs; Object (o): Access to object-level APIs for blobs, queue messages, table entities, and files. Possible values include: 'ResourceTypesS', 'ResourceTypesC', 'ResourceTypesO'
	ResourceTypes ResourceTypes `json:"signedResourceTypes,omitempty"`
	// Permissions - The signed permissions for the account SAS. Possible values include: Read (r), Write (w), Delete (d), List (l), Add (a), Create (c), Update (u) and Process (p). Possible values include: 'R', 'D', 'W', 'L', 'A', 'C', 'U', 'P'
	Permissions Permissions `json:"signedPermission,omitempty"`
	// IPAddressOrRange - An IP address or a range of IP addresses from which to accept requests.
	IPAddressOrRange *string `json:"signedIp,omitempty"`
	// Protocols - The protocol permitted for a request made with the account SAS. Possible values include: 'Httpshttp', 'HTTPS'
	Protocols HTTPProtocol `json:"signedProtocol,omitempty"`
	// SharedAccessStartTime - The time at which the SAS becomes valid.
	SharedAccessStartTime *date.Time `json:"signedStart,omitempty"`
	// SharedAccessExpiryTime - The time at which the shared access signature becomes invalid.
	SharedAccessExpiryTime *date.Time `json:"signedExpiry,omitempty"`
	// KeyToSign - The key to sign the account SAS token with.
	KeyToSign *string `json:"keyToSign,omitempty"`
}

// AccountsCreateFuture an abstraction for monitoring and retrieving the results of a long-running operation.
type AccountsCreateFuture struct {
	azure.Future
	req *http.Request
}

// Result returns the result of the asynchronous operation.
// If the operation has not completed it will return an error.
func (future AccountsCreateFuture) Result(client AccountsClient) (a Account, err error) {
	var done bool
	done, err = future.Done(client)
	if err != nil {
		return
	}
	if !done {
		return a, autorest.NewError("storage.AccountsCreateFuture", "Result", "asynchronous operation has not completed")
	}
	if future.PollingMethod() == azure.PollingLocation {
		a, err = client.CreateResponder(future.Response())
		return
	}
	var resp *http.Response
	resp, err = autorest.SendWithSender(client, autorest.ChangeToGet(future.req),
		autorest.DoRetryForStatusCodes(client.RetryAttempts, client.RetryDuration, autorest.StatusCodesForRetry...))
	if err != nil {
		return
	}
	a, err = client.CreateResponder(resp)
	return
}

// AccountUpdateParameters the parameters that can be provided when updating the storage account properties.
type AccountUpdateParameters struct {
	// Sku - Gets or sets the SKU name. Note that the SKU name cannot be updated to Standard_ZRS or Premium_LRS, nor can accounts of those sku names be updated to any other value.
	Sku *Sku `json:"sku,omitempty"`
	// Tags - Gets or sets a list of key value pairs that describe the resource. These tags can be used in viewing and grouping this resource (across resource groups). A maximum of 15 tags can be provided for a resource. Each tag must have a key no greater in length than 128 characters and a value no greater in length than 256 characters.
	Tags                               *map[string]*string `json:"tags,omitempty"`
	*AccountPropertiesUpdateParameters `json:"properties,omitempty"`
}

// UnmarshalJSON is the custom unmarshaler for AccountUpdateParameters struct.
func (aup *AccountUpdateParameters) UnmarshalJSON(body []byte) error {
	var m map[string]*json.RawMessage
	err := json.Unmarshal(body, &m)
	if err != nil {
		return err
	}
	var v *json.RawMessage

	v = m["sku"]
	if v != nil {
		var sku Sku
		err = json.Unmarshal(*m["sku"], &sku)
		if err != nil {
			return err
		}
		aup.Sku = &sku
	}

	v = m["tags"]
	if v != nil {
		var tags map[string]*string
		err = json.Unmarshal(*m["tags"], &tags)
		if err != nil {
			return err
		}
		aup.Tags = &tags
	}

	v = m["properties"]
	if v != nil {
		var properties AccountPropertiesUpdateParameters
		err = json.Unmarshal(*m["properties"], &properties)
		if err != nil {
			return err
		}
		aup.AccountPropertiesUpdateParameters = &properties
	}

	return nil
}

// CheckNameAvailabilityResult the CheckNameAvailability operation response.
type CheckNameAvailabilityResult struct {
	autorest.Response `json:"-"`
	// NameAvailable - Gets a boolean value that indicates whether the name is available for you to use. If true, the name is available. If false, the name has already been taken or is invalid and cannot be used.
	NameAvailable *bool `json:"nameAvailable,omitempty"`
	// Reason - Gets the reason that a storage account name could not be used. The Reason element is only returned if NameAvailable is false. Possible values include: 'AccountNameInvalid', 'AlreadyExists'
	Reason Reason `json:"reason,omitempty"`
	// Message - Gets an error message explaining the Reason value in more detail.
	Message *string `json:"message,omitempty"`
}

// CustomDomain the custom domain assigned to this storage account. This can be set via Update.
type CustomDomain struct {
	// Name - Gets or sets the custom domain name assigned to the storage account. Name is the CNAME source.
	Name *string `json:"name,omitempty"`
	// UseSubDomain - Indicates whether indirect CName validation is enabled. Default value is false. This should only be set on updates.
	UseSubDomain *bool `json:"useSubDomain,omitempty"`
}

// Encryption the encryption settings on the storage account.
type Encryption struct {
	// Services - List of services which support encryption.
	Services *EncryptionServices `json:"services,omitempty"`
	// KeySource - The encryption keySource (provider). Possible values (case-insensitive):  Microsoft.Storage
	KeySource *string `json:"keySource,omitempty"`
}

// EncryptionService a service that allows server-side encryption to be used.
type EncryptionService struct {
	// Enabled - A boolean indicating whether or not the service encrypts the data as it is stored.
	Enabled *bool `json:"enabled,omitempty"`
	// LastEnabledTime - Gets a rough estimate of the date/time when the encryption was last enabled by the user. Only returned when encryption is enabled. There might be some unencrypted blobs which were written after this time, as it is just a rough estimate.
	LastEnabledTime *date.Time `json:"lastEnabledTime,omitempty"`
}

// EncryptionServices a list of services that support encryption.
type EncryptionServices struct {
	// Blob - The encryption function of the blob storage service.
	Blob *EncryptionService `json:"blob,omitempty"`
	// File - The encryption function of the file storage service.
	File *EncryptionService `json:"file,omitempty"`
	// Table - The encryption function of the table storage service.
	Table *EncryptionService `json:"table,omitempty"`
	// Queue - The encryption function of the queue storage service.
	Queue *EncryptionService `json:"queue,omitempty"`
}

// Endpoints the URIs that are used to perform a retrieval of a public blob, queue, or table object.
type Endpoints struct {
	// Blob - Gets the blob endpoint.
	Blob *string `json:"blob,omitempty"`
	// Queue - Gets the queue endpoint.
	Queue *string `json:"queue,omitempty"`
	// Table - Gets the table endpoint.
	Table *string `json:"table,omitempty"`
	// File - Gets the file endpoint.
	File *string `json:"file,omitempty"`
}

// ListAccountSasResponse the List SAS credentials operation response.
type ListAccountSasResponse struct {
	autorest.Response `json:"-"`
	// AccountSasToken - List SAS credentials of storage account.
	AccountSasToken *string `json:"accountSasToken,omitempty"`
}

// ListServiceSasResponse the List service SAS credentials operation response.
type ListServiceSasResponse struct {
	autorest.Response `json:"-"`
	// ServiceSasToken - List service SAS credentials of speicific resource.
	ServiceSasToken *string `json:"serviceSasToken,omitempty"`
}

// Resource describes a storage resource.
type Resource struct {
	// ID - Resource Id
	ID *string `json:"id,omitempty"`
	// Name - Resource name
	Name *string `json:"name,omitempty"`
	// Type - Resource type
	Type *string `json:"type,omitempty"`
	// Location - Resource location
	Location *string `json:"location,omitempty"`
	// Tags - Tags assigned to a resource; can be used for viewing and grouping a resource (across resource groups).
	Tags *map[string]*string `json:"tags,omitempty"`
}

// ServiceSasParameters the parameters to list service SAS credentials of a speicific resource.
type ServiceSasParameters struct {
	// CanonicalizedResource - The canonical path to the signed resource.
	CanonicalizedResource *string `json:"canonicalizedResource,omitempty"`
	// Resource - The signed services accessible with the service SAS. Possible values include: Blob (b), Container (c), File (f), Share (s). Possible values include: 'ResourceEnumB', 'ResourceEnumC', 'ResourceEnumF', 'ResourceEnumS'
	Resource Resource `json:"signedResource,omitempty"`
	// Permissions - The signed permissions for the service SAS. Possible values include: Read (r), Write (w), Delete (d), List (l), Add (a), Create (c), Update (u) and Process (p). Possible values include: 'Permissions1R', 'Permissions1D', 'Permissions1W', 'Permissions1L', 'Permissions1A', 'Permissions1C', 'Permissions1U', 'Permissions1P'
	Permissions Permissions `json:"signedPermission,omitempty"`
	// IPAddressOrRange - An IP address or a range of IP addresses from which to accept requests.
	IPAddressOrRange *string `json:"signedIp,omitempty"`
	// Protocols - The protocol permitted for a request made with the account SAS. Possible values include: 'Httpshttp', 'HTTPS'
	Protocols HTTPProtocol `json:"signedProtocol,omitempty"`
	// SharedAccessStartTime - The time at which the SAS becomes valid.
	SharedAccessStartTime *date.Time `json:"signedStart,omitempty"`
	// SharedAccessExpiryTime - The time at which the shared access signature becomes invalid.
	SharedAccessExpiryTime *date.Time `json:"signedExpiry,omitempty"`
	// Identifier - A unique value up to 64 characters in length that correlates to an access policy specified for the container, queue, or table.
	Identifier *string `json:"signedIdentifier,omitempty"`
	// PartitionKeyStart - The start of partition key.
	PartitionKeyStart *string `json:"startPk,omitempty"`
	// PartitionKeyEnd - The end of partition key.
	PartitionKeyEnd *string `json:"endPk,omitempty"`
	// RowKeyStart - The start of row key.
	RowKeyStart *string `json:"startRk,omitempty"`
	// RowKeyEnd - The end of row key.
	RowKeyEnd *string `json:"endRk,omitempty"`
	// KeyToSign - The key to sign the account SAS token with.
	KeyToSign *string `json:"keyToSign,omitempty"`
	// CacheControl - The response header override for cache control.
	CacheControl *string `json:"rscc,omitempty"`
	// ContentDisposition - The response header override for content disposition.
	ContentDisposition *string `json:"rscd,omitempty"`
	// ContentEncoding - The response header override for content encoding.
	ContentEncoding *string `json:"rsce,omitempty"`
	// ContentLanguage - The response header override for content language.
	ContentLanguage *string `json:"rscl,omitempty"`
	// ContentType - The response header override for content type.
	ContentType *string `json:"rsct,omitempty"`
}

// Sku the SKU of the storage account.
type Sku struct {
	// Name - Gets or sets the sku name. Required for account creation; optional for update. Note that in older versions, sku name was called accountType. Possible values include: 'StandardLRS', 'StandardGRS', 'StandardRAGRS', 'StandardZRS', 'PremiumLRS'
	Name SkuName `json:"name,omitempty"`
	// Tier - Gets the sku tier. This is based on the SKU name. Possible values include: 'Standard', 'Premium'
	Tier SkuTier `json:"tier,omitempty"`
}

// Usage describes Storage Resource Usage.
type Usage struct {
	// Unit - Gets the unit of measurement. Possible values include: 'Count', 'Bytes', 'Seconds', 'Percent', 'CountsPerSecond', 'BytesPerSecond'
	Unit UsageUnit `json:"unit,omitempty"`
	// CurrentValue - Gets the current count of the allocated resources in the subscription.
	CurrentValue *int32 `json:"currentValue,omitempty"`
	// Limit - Gets the maximum count of the resources that can be allocated in the subscription.
	Limit *int32 `json:"limit,omitempty"`
	// Name - Gets the name of the type of usage.
	Name *UsageName `json:"name,omitempty"`
}

// UsageListResult the response from the List Usages operation.
type UsageListResult struct {
	autorest.Response `json:"-"`
	// Value - Gets or sets the list of Storage Resource Usages.
	Value *[]Usage `json:"value,omitempty"`
}

// UsageName the usage names that can be used; currently limited to StorageAccount.
type UsageName struct {
	// Value - Gets a string describing the resource name.
	Value *string `json:"value,omitempty"`
	// LocalizedValue - Gets a localized string describing the resource name.
	LocalizedValue *string `json:"localizedValue,omitempty"`
}
