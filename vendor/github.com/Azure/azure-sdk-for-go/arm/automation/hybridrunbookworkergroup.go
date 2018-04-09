package automation

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
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/validation"
	"net/http"
)

// HybridRunbookWorkerGroupClient is the automation Client
type HybridRunbookWorkerGroupClient struct {
	ManagementClient
}

// NewHybridRunbookWorkerGroupClient creates an instance of the HybridRunbookWorkerGroupClient client.
func NewHybridRunbookWorkerGroupClient(subscriptionID string) HybridRunbookWorkerGroupClient {
	return NewHybridRunbookWorkerGroupClientWithBaseURI(DefaultBaseURI, subscriptionID)
}

// NewHybridRunbookWorkerGroupClientWithBaseURI creates an instance of the HybridRunbookWorkerGroupClient client.
func NewHybridRunbookWorkerGroupClientWithBaseURI(baseURI string, subscriptionID string) HybridRunbookWorkerGroupClient {
	return HybridRunbookWorkerGroupClient{NewWithBaseURI(baseURI, subscriptionID)}
}

// Delete delete a hybrid runbook worker group.
//
// resourceGroupName is the resource group name. automationAccountName is automation account name.
// hybridRunbookWorkerGroupName is the hybrid runbook worker group name
func (client HybridRunbookWorkerGroupClient) Delete(resourceGroupName string, automationAccountName string, hybridRunbookWorkerGroupName string) (result autorest.Response, err error) {
	if err := validation.Validate([]validation.Validation{
		{TargetValue: resourceGroupName,
			Constraints: []validation.Constraint{{Target: "resourceGroupName", Name: validation.Pattern, Rule: `^[-\w\._]+$`, Chain: nil}}}}); err != nil {
		return result, validation.NewErrorWithValidationError(err, "automation.HybridRunbookWorkerGroupClient", "Delete")
	}

	req, err := client.DeletePreparer(resourceGroupName, automationAccountName, hybridRunbookWorkerGroupName)
	if err != nil {
		err = autorest.NewErrorWithError(err, "automation.HybridRunbookWorkerGroupClient", "Delete", nil, "Failure preparing request")
		return
	}

	resp, err := client.DeleteSender(req)
	if err != nil {
		result.Response = resp
		err = autorest.NewErrorWithError(err, "automation.HybridRunbookWorkerGroupClient", "Delete", resp, "Failure sending request")
		return
	}

	result, err = client.DeleteResponder(resp)
	if err != nil {
		err = autorest.NewErrorWithError(err, "automation.HybridRunbookWorkerGroupClient", "Delete", resp, "Failure responding to request")
	}

	return
}

// DeletePreparer prepares the Delete request.
func (client HybridRunbookWorkerGroupClient) DeletePreparer(resourceGroupName string, automationAccountName string, hybridRunbookWorkerGroupName string) (*http.Request, error) {
	pathParameters := map[string]interface{}{
		"automationAccountName":        autorest.Encode("path", automationAccountName),
		"hybridRunbookWorkerGroupName": autorest.Encode("path", hybridRunbookWorkerGroupName),
		"resourceGroupName":            autorest.Encode("path", resourceGroupName),
		"subscriptionId":               autorest.Encode("path", client.SubscriptionID),
	}

	const APIVersion = "2015-10-31"
	queryParameters := map[string]interface{}{
		"api-version": APIVersion,
	}

	preparer := autorest.CreatePreparer(
		autorest.AsDelete(),
		autorest.WithBaseURL(client.BaseURI),
		autorest.WithPathParameters("/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.Automation/automationAccounts/{automationAccountName}/hybridRunbookWorkerGroups/{hybridRunbookWorkerGroupName}", pathParameters),
		autorest.WithQueryParameters(queryParameters))
	return preparer.Prepare(&http.Request{})
}

// DeleteSender sends the Delete request. The method will close the
// http.Response Body if it receives an error.
func (client HybridRunbookWorkerGroupClient) DeleteSender(req *http.Request) (*http.Response, error) {
	return autorest.SendWithSender(client,
		req,
		azure.DoRetryWithRegistration(client.Client))
}

// DeleteResponder handles the response to the Delete request. The method always
// closes the http.Response Body.
func (client HybridRunbookWorkerGroupClient) DeleteResponder(resp *http.Response) (result autorest.Response, err error) {
	err = autorest.Respond(
		resp,
		client.ByInspecting(),
		azure.WithErrorUnlessStatusCode(http.StatusOK),
		autorest.ByClosing())
	result.Response = resp
	return
}

// Get retrieve a hybrid runbook worker group.
//
// resourceGroupName is the resource group name. automationAccountName is the automation account name.
// hybridRunbookWorkerGroupName is the hybrid runbook worker group name
func (client HybridRunbookWorkerGroupClient) Get(resourceGroupName string, automationAccountName string, hybridRunbookWorkerGroupName string) (result HybridRunbookWorkerGroup, err error) {
	if err := validation.Validate([]validation.Validation{
		{TargetValue: resourceGroupName,
			Constraints: []validation.Constraint{{Target: "resourceGroupName", Name: validation.Pattern, Rule: `^[-\w\._]+$`, Chain: nil}}}}); err != nil {
		return result, validation.NewErrorWithValidationError(err, "automation.HybridRunbookWorkerGroupClient", "Get")
	}

	req, err := client.GetPreparer(resourceGroupName, automationAccountName, hybridRunbookWorkerGroupName)
	if err != nil {
		err = autorest.NewErrorWithError(err, "automation.HybridRunbookWorkerGroupClient", "Get", nil, "Failure preparing request")
		return
	}

	resp, err := client.GetSender(req)
	if err != nil {
		result.Response = autorest.Response{Response: resp}
		err = autorest.NewErrorWithError(err, "automation.HybridRunbookWorkerGroupClient", "Get", resp, "Failure sending request")
		return
	}

	result, err = client.GetResponder(resp)
	if err != nil {
		err = autorest.NewErrorWithError(err, "automation.HybridRunbookWorkerGroupClient", "Get", resp, "Failure responding to request")
	}

	return
}

// GetPreparer prepares the Get request.
func (client HybridRunbookWorkerGroupClient) GetPreparer(resourceGroupName string, automationAccountName string, hybridRunbookWorkerGroupName string) (*http.Request, error) {
	pathParameters := map[string]interface{}{
		"automationAccountName":        autorest.Encode("path", automationAccountName),
		"hybridRunbookWorkerGroupName": autorest.Encode("path", hybridRunbookWorkerGroupName),
		"resourceGroupName":            autorest.Encode("path", resourceGroupName),
		"subscriptionId":               autorest.Encode("path", client.SubscriptionID),
	}

	const APIVersion = "2015-10-31"
	queryParameters := map[string]interface{}{
		"api-version": APIVersion,
	}

	preparer := autorest.CreatePreparer(
		autorest.AsGet(),
		autorest.WithBaseURL(client.BaseURI),
		autorest.WithPathParameters("/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.Automation/automationAccounts/{automationAccountName}/hybridRunbookWorkerGroups/{hybridRunbookWorkerGroupName}", pathParameters),
		autorest.WithQueryParameters(queryParameters))
	return preparer.Prepare(&http.Request{})
}

// GetSender sends the Get request. The method will close the
// http.Response Body if it receives an error.
func (client HybridRunbookWorkerGroupClient) GetSender(req *http.Request) (*http.Response, error) {
	return autorest.SendWithSender(client,
		req,
		azure.DoRetryWithRegistration(client.Client))
}

// GetResponder handles the response to the Get request. The method always
// closes the http.Response Body.
func (client HybridRunbookWorkerGroupClient) GetResponder(resp *http.Response) (result HybridRunbookWorkerGroup, err error) {
	err = autorest.Respond(
		resp,
		client.ByInspecting(),
		azure.WithErrorUnlessStatusCode(http.StatusOK),
		autorest.ByUnmarshallingJSON(&result),
		autorest.ByClosing())
	result.Response = autorest.Response{Response: resp}
	return
}

// ListByAutomationAccount retrieve a list of hybrid runbook worker groups.
//
// resourceGroupName is the resource group name. automationAccountName is the automation account name.
func (client HybridRunbookWorkerGroupClient) ListByAutomationAccount(resourceGroupName string, automationAccountName string) (result HybridRunbookWorkerGroupsListResult, err error) {
	if err := validation.Validate([]validation.Validation{
		{TargetValue: resourceGroupName,
			Constraints: []validation.Constraint{{Target: "resourceGroupName", Name: validation.Pattern, Rule: `^[-\w\._]+$`, Chain: nil}}}}); err != nil {
		return result, validation.NewErrorWithValidationError(err, "automation.HybridRunbookWorkerGroupClient", "ListByAutomationAccount")
	}

	req, err := client.ListByAutomationAccountPreparer(resourceGroupName, automationAccountName)
	if err != nil {
		err = autorest.NewErrorWithError(err, "automation.HybridRunbookWorkerGroupClient", "ListByAutomationAccount", nil, "Failure preparing request")
		return
	}

	resp, err := client.ListByAutomationAccountSender(req)
	if err != nil {
		result.Response = autorest.Response{Response: resp}
		err = autorest.NewErrorWithError(err, "automation.HybridRunbookWorkerGroupClient", "ListByAutomationAccount", resp, "Failure sending request")
		return
	}

	result, err = client.ListByAutomationAccountResponder(resp)
	if err != nil {
		err = autorest.NewErrorWithError(err, "automation.HybridRunbookWorkerGroupClient", "ListByAutomationAccount", resp, "Failure responding to request")
	}

	return
}

// ListByAutomationAccountPreparer prepares the ListByAutomationAccount request.
func (client HybridRunbookWorkerGroupClient) ListByAutomationAccountPreparer(resourceGroupName string, automationAccountName string) (*http.Request, error) {
	pathParameters := map[string]interface{}{
		"automationAccountName": autorest.Encode("path", automationAccountName),
		"resourceGroupName":     autorest.Encode("path", resourceGroupName),
		"subscriptionId":        autorest.Encode("path", client.SubscriptionID),
	}

	const APIVersion = "2015-10-31"
	queryParameters := map[string]interface{}{
		"api-version": APIVersion,
	}

	preparer := autorest.CreatePreparer(
		autorest.AsGet(),
		autorest.WithBaseURL(client.BaseURI),
		autorest.WithPathParameters("/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.Automation/automationAccounts/{automationAccountName}/hybridRunbookWorkerGroups", pathParameters),
		autorest.WithQueryParameters(queryParameters))
	return preparer.Prepare(&http.Request{})
}

// ListByAutomationAccountSender sends the ListByAutomationAccount request. The method will close the
// http.Response Body if it receives an error.
func (client HybridRunbookWorkerGroupClient) ListByAutomationAccountSender(req *http.Request) (*http.Response, error) {
	return autorest.SendWithSender(client,
		req,
		azure.DoRetryWithRegistration(client.Client))
}

// ListByAutomationAccountResponder handles the response to the ListByAutomationAccount request. The method always
// closes the http.Response Body.
func (client HybridRunbookWorkerGroupClient) ListByAutomationAccountResponder(resp *http.Response) (result HybridRunbookWorkerGroupsListResult, err error) {
	err = autorest.Respond(
		resp,
		client.ByInspecting(),
		azure.WithErrorUnlessStatusCode(http.StatusOK),
		autorest.ByUnmarshallingJSON(&result),
		autorest.ByClosing())
	result.Response = autorest.Response{Response: resp}
	return
}

// ListByAutomationAccountNextResults retrieves the next set of results, if any.
func (client HybridRunbookWorkerGroupClient) ListByAutomationAccountNextResults(lastResults HybridRunbookWorkerGroupsListResult) (result HybridRunbookWorkerGroupsListResult, err error) {
	req, err := lastResults.HybridRunbookWorkerGroupsListResultPreparer()
	if err != nil {
		return result, autorest.NewErrorWithError(err, "automation.HybridRunbookWorkerGroupClient", "ListByAutomationAccount", nil, "Failure preparing next results request")
	}
	if req == nil {
		return
	}

	resp, err := client.ListByAutomationAccountSender(req)
	if err != nil {
		result.Response = autorest.Response{Response: resp}
		return result, autorest.NewErrorWithError(err, "automation.HybridRunbookWorkerGroupClient", "ListByAutomationAccount", resp, "Failure sending next results request")
	}

	result, err = client.ListByAutomationAccountResponder(resp)
	if err != nil {
		err = autorest.NewErrorWithError(err, "automation.HybridRunbookWorkerGroupClient", "ListByAutomationAccount", resp, "Failure responding to next results request")
	}

	return
}

// ListByAutomationAccountComplete gets all elements from the list without paging.
func (client HybridRunbookWorkerGroupClient) ListByAutomationAccountComplete(resourceGroupName string, automationAccountName string, cancel <-chan struct{}) (<-chan HybridRunbookWorkerGroup, <-chan error) {
	resultChan := make(chan HybridRunbookWorkerGroup)
	errChan := make(chan error, 1)
	go func() {
		defer func() {
			close(resultChan)
			close(errChan)
		}()
		list, err := client.ListByAutomationAccount(resourceGroupName, automationAccountName)
		if err != nil {
			errChan <- err
			return
		}
		if list.Value != nil {
			for _, item := range *list.Value {
				select {
				case <-cancel:
					return
				case resultChan <- item:
					// Intentionally left blank
				}
			}
		}
		for list.NextLink != nil {
			list, err = client.ListByAutomationAccountNextResults(list)
			if err != nil {
				errChan <- err
				return
			}
			if list.Value != nil {
				for _, item := range *list.Value {
					select {
					case <-cancel:
						return
					case resultChan <- item:
						// Intentionally left blank
					}
				}
			}
		}
	}()
	return resultChan, errChan
}

// Update update a hybrid runbook worker group.
//
// resourceGroupName is the resource group name. automationAccountName is the automation account name.
// hybridRunbookWorkerGroupName is the hybrid runbook worker group name parameters is the hybrid runbook worker group
func (client HybridRunbookWorkerGroupClient) Update(resourceGroupName string, automationAccountName string, hybridRunbookWorkerGroupName string, parameters HybridRunbookWorkerGroupUpdateParameters) (result HybridRunbookWorkerGroup, err error) {
	if err := validation.Validate([]validation.Validation{
		{TargetValue: resourceGroupName,
			Constraints: []validation.Constraint{{Target: "resourceGroupName", Name: validation.Pattern, Rule: `^[-\w\._]+$`, Chain: nil}}}}); err != nil {
		return result, validation.NewErrorWithValidationError(err, "automation.HybridRunbookWorkerGroupClient", "Update")
	}

	req, err := client.UpdatePreparer(resourceGroupName, automationAccountName, hybridRunbookWorkerGroupName, parameters)
	if err != nil {
		err = autorest.NewErrorWithError(err, "automation.HybridRunbookWorkerGroupClient", "Update", nil, "Failure preparing request")
		return
	}

	resp, err := client.UpdateSender(req)
	if err != nil {
		result.Response = autorest.Response{Response: resp}
		err = autorest.NewErrorWithError(err, "automation.HybridRunbookWorkerGroupClient", "Update", resp, "Failure sending request")
		return
	}

	result, err = client.UpdateResponder(resp)
	if err != nil {
		err = autorest.NewErrorWithError(err, "automation.HybridRunbookWorkerGroupClient", "Update", resp, "Failure responding to request")
	}

	return
}

// UpdatePreparer prepares the Update request.
func (client HybridRunbookWorkerGroupClient) UpdatePreparer(resourceGroupName string, automationAccountName string, hybridRunbookWorkerGroupName string, parameters HybridRunbookWorkerGroupUpdateParameters) (*http.Request, error) {
	pathParameters := map[string]interface{}{
		"automationAccountName":        autorest.Encode("path", automationAccountName),
		"hybridRunbookWorkerGroupName": autorest.Encode("path", hybridRunbookWorkerGroupName),
		"resourceGroupName":            autorest.Encode("path", resourceGroupName),
		"subscriptionId":               autorest.Encode("path", client.SubscriptionID),
	}

	const APIVersion = "2015-10-31"
	queryParameters := map[string]interface{}{
		"api-version": APIVersion,
	}

	preparer := autorest.CreatePreparer(
		autorest.AsJSON(),
		autorest.AsPatch(),
		autorest.WithBaseURL(client.BaseURI),
		autorest.WithPathParameters("/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.Automation/automationAccounts/{automationAccountName}/hybridRunbookWorkerGroups/{hybridRunbookWorkerGroupName}", pathParameters),
		autorest.WithJSON(parameters),
		autorest.WithQueryParameters(queryParameters))
	return preparer.Prepare(&http.Request{})
}

// UpdateSender sends the Update request. The method will close the
// http.Response Body if it receives an error.
func (client HybridRunbookWorkerGroupClient) UpdateSender(req *http.Request) (*http.Response, error) {
	return autorest.SendWithSender(client,
		req,
		azure.DoRetryWithRegistration(client.Client))
}

// UpdateResponder handles the response to the Update request. The method always
// closes the http.Response Body.
func (client HybridRunbookWorkerGroupClient) UpdateResponder(resp *http.Response) (result HybridRunbookWorkerGroup, err error) {
	err = autorest.Respond(
		resp,
		client.ByInspecting(),
		azure.WithErrorUnlessStatusCode(http.StatusOK),
		autorest.ByUnmarshallingJSON(&result),
		autorest.ByClosing())
	result.Response = autorest.Response{Response: resp}
	return
}
