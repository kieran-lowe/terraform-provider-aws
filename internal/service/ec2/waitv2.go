// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func waitVPCCreatedV2(ctx context.Context, conn *ec2.Client, id string) (*types.Vpc, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.VpcStatePending),
		Target:  enum.Slice(types.VpcStateAvailable),
		Refresh: statusVPCStateV2(ctx, conn, id),
		Timeout: vpcCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.Vpc); ok {
		return output, err
	}

	return nil, err
}

func waitVPCIPv6CIDRBlockAssociationCreatedV2(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*types.VpcCidrBlockState, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(types.VpcCidrBlockStateCodeAssociating, types.VpcCidrBlockStateCodeDisassociated, types.VpcCidrBlockStateCodeFailing),
		Target:     enum.Slice(types.VpcCidrBlockStateCodeAssociated),
		Refresh:    statusVPCIPv6CIDRBlockAssociationStateV2(ctx, conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.VpcCidrBlockState); ok {
		if state := output.State; state == types.VpcCidrBlockStateCodeFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusMessage)))
		}

		return output, err
	}

	return nil, err
}

func waitVPCAttributeUpdatedV2(ctx context.Context, conn *ec2.Client, vpcID string, attribute types.VpcAttributeName, expectedValue bool) (*types.Vpc, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Target:     []string{strconv.FormatBool(expectedValue)},
		Refresh:    statusVPCAttributeValueV2(ctx, conn, vpcID, attribute),
		Timeout:    ec2PropagationTimeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.Vpc); ok {
		return output, err
	}

	return nil, err
}

func waitVPCIPv6CIDRBlockAssociationDeletedV2(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*types.VpcCidrBlockState, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(types.VpcCidrBlockStateCodeAssociated, types.VpcCidrBlockStateCodeDisassociating, types.VpcCidrBlockStateCodeFailing),
		Target:     []string{},
		Refresh:    statusVPCIPv6CIDRBlockAssociationStateV2(ctx, conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.VpcCidrBlockState); ok {
		if state := output.State; state == types.VpcCidrBlockStateCodeFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusMessage)))
		}

		return output, err
	}

	return nil, err
}

func waitNetworkInterfaceAvailableAfterUseV2(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*types.NetworkInterface, error) {
	// Hyperplane attached ENI.
	// Wait for it to be moved into a removable state.
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(types.NetworkInterfaceStatusInUse),
		Target:     enum.Slice(types.NetworkInterfaceStatusAvailable),
		Timeout:    timeout,
		Refresh:    statusNetworkInterfaceV2(ctx, conn, id),
		Delay:      10 * time.Second,
		MinTimeout: 10 * time.Second,
		// Handle EC2 ENI eventual consistency. It can take up to 3 minutes.
		ContinuousTargetOccurence: 18,
		NotFoundChecks:            1,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.NetworkInterface); ok {
		return output, err
	}

	return nil, err
}

func waitNetworkInterfaceCreatedV2(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*types.NetworkInterface, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{NetworkInterfaceStatusPending},
		Target:  enum.Slice(types.NetworkInterfaceStatusAvailable),
		Timeout: timeout,
		Refresh: statusNetworkInterfaceV2(ctx, conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.NetworkInterface); ok {
		return output, err
	}

	return nil, err
}

func waitNetworkInterfaceAttachedV2(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*types.NetworkInterfaceAttachment, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.AttachmentStatusAttaching),
		Target:  enum.Slice(types.AttachmentStatusAttached),
		Timeout: timeout,
		Refresh: statusNetworkInterfaceAttachmentV2(ctx, conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.NetworkInterfaceAttachment); ok {
		return output, err
	}

	return nil, err
}

func waitNetworkInterfaceDetachedV2(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*types.NetworkInterfaceAttachment, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.AttachmentStatusAttached, types.AttachmentStatusDetaching),
		Target:  enum.Slice(types.AttachmentStatusDetached),
		Timeout: timeout,
		Refresh: statusNetworkInterfaceAttachmentV2(ctx, conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.NetworkInterfaceAttachment); ok {
		return output, err
	}

	return nil, err
}

func waitVPCEndpointAcceptedV2(ctx context.Context, conn *ec2.Client, vpcEndpointID string, timeout time.Duration) (*types.VpcEndpoint, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(vpcEndpointStatePendingAcceptance),
		Target:     enum.Slice(vpcEndpointStateAvailable),
		Timeout:    timeout,
		Refresh:    statusVPCEndpointStateV2(ctx, conn, vpcEndpointID),
		Delay:      5 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.VpcEndpoint); ok {
		if state, lastError := output.State, output.LastError; state == types.StateFailed && lastError != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.ToString(lastError.Code), aws.ToString(lastError.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitVPCEndpointAvailableV2(ctx context.Context, conn *ec2.Client, vpcEndpointID string, timeout time.Duration) (*types.VpcEndpoint, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(vpcEndpointStatePending),
		Target:     enum.Slice(vpcEndpointStateAvailable, vpcEndpointStatePendingAcceptance),
		Timeout:    timeout,
		Refresh:    statusVPCEndpointStateV2(ctx, conn, vpcEndpointID),
		Delay:      5 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.VpcEndpoint); ok {
		if state, lastError := output.State, output.LastError; state == types.StateFailed && lastError != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.ToString(lastError.Code), aws.ToString(lastError.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitVPCEndpointDeletedV2(ctx context.Context, conn *ec2.Client, vpcEndpointID string, timeout time.Duration) (*types.VpcEndpoint, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(vpcEndpointStateDeleting, vpcEndpointStateDeleted),
		Target:     []string{},
		Refresh:    statusVPCEndpointStateV2(ctx, conn, vpcEndpointID),
		Timeout:    timeout,
		Delay:      5 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.VpcEndpoint); ok {
		return output, err
	}

	return nil, err
}

func waitRouteDeletedV2(ctx context.Context, conn *ec2.Client, routeFinder routeFinderV2, routeTableID, destination string, timeout time.Duration) (*types.Route, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{RouteStatusReady},
		Target:                    []string{},
		Refresh:                   statusRouteV2(ctx, conn, routeFinder, routeTableID, destination),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.Route); ok {
		return output, err
	}

	return nil, err
}

func waitRouteReadyV2(ctx context.Context, conn *ec2.Client, routeFinder routeFinderV2, routeTableID, destination string, timeout time.Duration) (*types.Route, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{RouteStatusReady},
		Refresh:                   statusRouteV2(ctx, conn, routeFinder, routeTableID, destination),
		Timeout:                   timeout,
		NotFoundChecks:            RouteNotFoundChecks,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.Route); ok {
		return output, err
	}

	return nil, err
}

func waitRouteTableReadyV2(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*types.RouteTable, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{RouteTableStatusReady},
		Refresh:                   statusRouteTableV2(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            RouteTableNotFoundChecks,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.RouteTable); ok {
		return output, err
	}

	return nil, err
}

func waitRouteTableDeletedV2(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*types.RouteTable, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{RouteTableStatusReady},
		Target:                    []string{},
		Refresh:                   statusRouteTableV2(ctx, conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.RouteTable); ok {
		return output, err
	}

	return nil, err
}

func waitRouteTableAssociationCreatedV2(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*types.RouteTableAssociationState, error) {
	stateConf := &retry.StateChangeConf{
		Pending:        enum.Slice(types.RouteTableAssociationStateCodeAssociating),
		Target:         enum.Slice(types.RouteTableAssociationStateCodeAssociated),
		Refresh:        statusRouteTableAssociationStateV2(ctx, conn, id),
		Timeout:        timeout,
		NotFoundChecks: RouteTableAssociationCreatedNotFoundChecks,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.RouteTableAssociationState); ok {
		if output.State == types.RouteTableAssociationStateCodeFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusMessage)))
		}

		return output, err
	}

	return nil, err
}

func waitRouteTableAssociationDeletedV2(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*types.RouteTableAssociationState, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.RouteTableAssociationStateCodeDisassociating, types.RouteTableAssociationStateCodeAssociated),
		Target:  []string{},
		Refresh: statusRouteTableAssociationStateV2(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.RouteTableAssociationState); ok {
		if output.State == types.RouteTableAssociationStateCodeFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusMessage)))
		}

		return output, err
	}

	return nil, err
}

func waitRouteTableAssociationUpdatedV2(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*types.RouteTableAssociationState, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.RouteTableAssociationStateCodeAssociating),
		Target:  enum.Slice(types.RouteTableAssociationStateCodeAssociated),
		Refresh: statusRouteTableAssociationStateV2(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.RouteTableAssociationState); ok {
		if output.State == types.RouteTableAssociationStateCodeFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusMessage)))
		}

		return output, err
	}

	return nil, err
}

func waitVPCEndpointServiceAvailableV2(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*types.ServiceConfiguration, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(types.ServiceStatePending),
		Target:     enum.Slice(types.ServiceStateAvailable),
		Refresh:    statusVPCEndpointServiceStateAvailableV2(ctx, conn, id),
		Timeout:    timeout,
		Delay:      5 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.ServiceConfiguration); ok {
		return output, err
	}

	return nil, err
}

func waitVPCEndpointServiceDeletedV2(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*types.ServiceConfiguration, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(types.ServiceStateAvailable, types.ServiceStateDeleting),
		Target:     []string{},
		Timeout:    timeout,
		Refresh:    statusVPCEndpointServiceStateDeletedV2(ctx, conn, id),
		Delay:      5 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.ServiceConfiguration); ok {
		return output, err
	}

	return nil, err
}

func waitVPCEndpointRouteTableAssociationReadyV2(ctx context.Context, conn *ec2.Client, vpcEndpointID, routeTableID string) error {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    enum.Slice(VPCEndpointRouteTableAssociationStatusReady),
		Refresh:                   statusVPCEndpointRouteTableAssociationV2(ctx, conn, vpcEndpointID, routeTableID),
		Timeout:                   ec2PropagationTimeout,
		ContinuousTargetOccurence: 2,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func waitVPCEndpointRouteTableAssociationDeletedV2(ctx context.Context, conn *ec2.Client, vpcEndpointID, routeTableID string) error {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(VPCEndpointRouteTableAssociationStatusReady),
		Target:                    []string{},
		Refresh:                   statusVPCEndpointRouteTableAssociationV2(ctx, conn, vpcEndpointID, routeTableID),
		Timeout:                   ec2PropagationTimeout,
		ContinuousTargetOccurence: 2,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func waitVPCEndpointConnectionAcceptedV2(ctx context.Context, conn *ec2.Client, serviceID, vpcEndpointID string, timeout time.Duration) (*types.VpcEndpointConnection, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{vpcEndpointStatePendingAcceptance, vpcEndpointStatePending},
		Target:     []string{vpcEndpointStateAvailable},
		Refresh:    statusVPCEndpointConnectionVPCEndpointStateV2(ctx, conn, serviceID, vpcEndpointID),
		Timeout:    timeout,
		Delay:      5 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.VpcEndpointConnection); ok {
		return output, err
	}

	return nil, err
}

func waitVPCEndpointServicePrivateDNSNameVerifiedV2(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*types.PrivateDnsNameConfiguration, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(types.DnsNameStatePendingVerification),
		Target:                    enum.Slice(types.DnsNameStateVerified),
		Refresh:                   statusVPCEndpointServicePrivateDNSNameConfigurationV2(ctx, conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*types.PrivateDnsNameConfiguration); ok {
		return out, err
	}

	return nil, err
}
