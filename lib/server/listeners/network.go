/*
 * Copyright 2018-2019, CS Systemes d'Information, http://www.c-s.fr
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package listeners

import (
	"context"
	"fmt"
	"github.com/asaskevich/govalidator"
	"github.com/sirupsen/logrus"

	googleprotobuf "github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/CS-SI/SafeScale/lib"
	"github.com/CS-SI/SafeScale/lib/server/handlers"
	"github.com/CS-SI/SafeScale/lib/server/iaas/resources"
	"github.com/CS-SI/SafeScale/lib/server/iaas/resources/enums/ipversion"
	conv "github.com/CS-SI/SafeScale/lib/server/utils"
	srvutils "github.com/CS-SI/SafeScale/lib/server/utils"
	"github.com/CS-SI/SafeScale/lib/utils"
	"github.com/CS-SI/SafeScale/lib/utils/concurrency"
	"github.com/CS-SI/SafeScale/lib/utils/scerr"
)

// safescale network create net1 --cidr="192.145.0.0/16" --cpu=2 --ram=7 --disk=100 --os="Ubuntu 16.04" (par défault "192.168.0.0/24", on crée une gateway sur chaque réseau: gw-net1)
// safescale network list
// safescale network delete net1
// safescale network inspect net1

// NetworkListener network service server grpc
type NetworkListener struct{}

// Create a new network
func (s *NetworkListener) Create(ctx context.Context, in *pb.NetworkDefinition) (_ *pb.Network, err error) {
	if s == nil {
		return nil, scerr.InvalidInstanceError().ToGRPCStatus()
	}
	if in == nil {
		return nil, scerr.InvalidParameterError("in", "cannot be nil").ToGRPCStatus()
	}
	if ctx == nil {
		return nil, scerr.InvalidParameterError("ctx", "cannot be nil").ToGRPCStatus()
	}

	ok, err := govalidator.ValidateStruct(in)
	if err == nil {
		if !ok {
			logrus.Warnf("Structure validation failure: %v", in) // FIXME Generate json tags in protobuf
		}
	}

	networkName := in.GetName()
	if networkName == "" {
		return nil, scerr.InvalidRequestError("cannot create network: name can't be empty string")
	}

	tracer := concurrency.NewTracer(nil, fmt.Sprintf("('%s')", networkName), true).WithStopwatch().GoingIn()
	defer tracer.OnExitTrace()()
	defer scerr.OnExitLogError(tracer.TraceMessage(""), &err)()

	ctx, cancelFunc := context.WithCancel(ctx)
	// LATER: handle jobregister error
	if err := srvutils.JobRegister(ctx, cancelFunc, "Create network "+networkName); err == nil {
		defer srvutils.JobDeregister(ctx)
	} /* else {
		return nil, scerr.InvalidInstanceContentError("ctx", "has no uuid").ToGRPCStatus()
	}*/

	tenant := GetCurrentTenant()
	if tenant == nil {
		msg := "cannot create network: no tenant set"
		tracer.Trace(utils.Capitalize(msg))
		return nil, status.Errorf(codes.FailedPrecondition, msg)
	}

	var (
		sizing    *resources.SizingRequirements
		gwImageID string
		gwName    string
	)
	if in.Gateway == nil || in.Gateway.Sizing == nil {
		sizing = &resources.SizingRequirements{
			MinCores:    int(in.Gateway.Sizing.MinCpuCount),
			MaxCores:    int(in.Gateway.Sizing.MaxCpuCount),
			MinRAMSize:  in.Gateway.Sizing.MinRamSize,
			MaxRAMSize:  in.Gateway.Sizing.MaxRamSize,
			MinDiskSize: int(in.Gateway.Sizing.MinDiskSize),
			MinGPU:      int(in.Gateway.Sizing.GpuCount),
			MinFreq:     in.Gateway.Sizing.MinCpuFreq,
		}
	} else {
		s := srvutils.FromPBHostSizing(*in.Gateway.Sizing)
		sizing = &s
	}
	if in.Gateway != nil {
		gwImageID = in.GetGateway().GetImageId()
		gwName = in.GetGateway().GetName()
	}

	handler := handlers.NewNetworkHandler(tenant.Service)
	network, err := handler.Create(ctx,
		networkName,
		in.GetCidr(),
		ipversion.IPv4,
		*sizing,
		gwImageID,
		gwName,
		in.FailOver,
	)
	if err != nil {
		// returned error is not wrap with "cannot create network" because already tells why the
		return nil, scerr.ToGRPCStatus(err)
	}

	tracer.Trace("Network '%s' successfuly created.", networkName)
	return conv.ToPBNetwork(network), nil
}

// List existing networks
func (s *NetworkListener) List(ctx context.Context, in *pb.NetworkListRequest) (_ *pb.NetworkList, err error) {
	if s == nil {
		return nil, scerr.InvalidInstanceError().ToGRPCStatus()
	}
	if in == nil {
		return nil, scerr.InvalidParameterError("in", "cannot be nil").ToGRPCStatus()
	}
	if ctx == nil {
		return nil, scerr.InvalidParameterError("ctx", "cannot be nil").ToGRPCStatus()
	}

	ok, err := govalidator.ValidateStruct(in)
	if err == nil {
		if !ok {
			logrus.Warnf("Structure validation failure: %v", in) // FIXME Generate json tags in protobuf
		}
	}

	tracer := concurrency.NewTracer(nil, "", true).WithStopwatch().GoingIn()
	defer tracer.OnExitTrace()()
	defer scerr.OnExitLogError(tracer.TraceMessage(""), &err)()

	ctx, cancelFunc := context.WithCancel(ctx)
	// LATER: handle jobregister error
	if err := srvutils.JobRegister(ctx, cancelFunc, "List networks"); err == nil {
		defer srvutils.JobDeregister(ctx)
	} /* else {
		return nil, scerr.InvalidInstanceContentError("ctx", "has no uuid").ToGRPCStatus()
	}*/

	tenant := GetCurrentTenant()
	if tenant == nil {
		msg := "cannot list networks: no tenant set"
		tracer.Trace(utils.Capitalize(msg))
		return nil, status.Errorf(codes.FailedPrecondition, msg)
	}

	handler := handlers.NewNetworkHandler(tenant.Service)
	networks, err := handler.List(ctx, in.GetAll())
	if err != nil {
		return nil, scerr.Wrap(err, "cannot list networks").ToGRPCStatus()
	}

	// Map resources.Network to pb.Network
	var pbnetworks []*pb.Network
	for _, network := range networks {
		pbnetworks = append(pbnetworks, conv.ToPBNetwork(network))
	}
	rv := &pb.NetworkList{Networks: pbnetworks}
	return rv, nil
}

// Inspect returns infos on a network
func (s *NetworkListener) Inspect(ctx context.Context, in *pb.Reference) (_ *pb.Network, err error) {
	if s == nil {
		return nil, scerr.InvalidInstanceError().ToGRPCStatus()
	}
	if in == nil {
		return nil, scerr.InvalidParameterError("in", "cannot be nil").ToGRPCStatus()
	}
	if ctx == nil {
		return nil, scerr.InvalidParameterError("ctx", "cannot be nil").ToGRPCStatus()
	}

	ok, err := govalidator.ValidateStruct(in)
	if err == nil {
		if !ok {
			logrus.Warnf("Structure validation failure: %v", in) // FIXME Generate json tags in protobuf
		}
	}

	ref := srvutils.GetReference(in)
	if ref == "" {
		return nil, scerr.InvalidRequestError("cannot inspect network: neither name nor id given as reference").ToGRPCStatus()
	}

	tracer := concurrency.NewTracer(nil, fmt.Sprintf("('%s')", ref), true).WithStopwatch().GoingIn()
	defer tracer.OnExitTrace()()
	defer scerr.OnExitLogError(tracer.TraceMessage(""), &err)()

	ctx, cancelFunc := context.WithCancel(ctx)
	// LATER: handle jobregister error
	if err := srvutils.JobRegister(ctx, cancelFunc, "Inspect network "+in.GetName()); err == nil {
		defer srvutils.JobDeregister(ctx)
	} /* else {
		return nil, scerr.InvalidInstanceContentError("ctx", "has no uuid").ToGRPCStatus()
	}*/

	tenant := GetCurrentTenant()
	if tenant == nil {
		msg := "cannot inspect network: no tenant set"
		tracer.Trace(utils.Capitalize(msg))
		return nil, status.Errorf(codes.FailedPrecondition, msg)
	}

	handler := handlers.NewNetworkHandler(currentTenant.Service)
	network, err := handler.Inspect(ctx, ref)
	if err != nil {
		return nil, scerr.Wrap(err, "cannot inspect network").ToGRPCStatus()
	}
	// this _must not_ happen, but InspectHost has different implementations for each stack, and sometimes mistakes happens, so the test is necessary
	if network == nil {
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("cannot inspect network '%s': not found", ref))
	}

	return conv.ToPBNetwork(network), nil
}

// Delete a network
func (s *NetworkListener) Delete(ctx context.Context, in *pb.Reference) (empty *googleprotobuf.Empty, err error) {
	empty = &googleprotobuf.Empty{}
	if s == nil {
		return empty, scerr.InvalidInstanceError().ToGRPCStatus()
	}
	if in == nil {
		return empty, scerr.InvalidParameterError("in", "cannot be nil").ToGRPCStatus()
	}
	if ctx == nil {
		return empty, scerr.InvalidParameterError("ctx", "cannot be nil").ToGRPCStatus()
	}

	ok, err := govalidator.ValidateStruct(in)
	if err == nil {
		if !ok {
			logrus.Warnf("Structure validation failure: %v", in) // FIXME Generate json tags in protobuf
		}
	}

	ref := srvutils.GetReference(in)
	if ref == "" {
		return empty, scerr.InvalidRequestError("cannot delete network: neither name nor id given as reference")
	}

	tracer := concurrency.NewTracer(nil, fmt.Sprintf("('%s')", ref), true).WithStopwatch().GoingIn()
	defer tracer.OnExitTrace()()
	defer scerr.OnExitLogError(tracer.TraceMessage(""), &err)()

	ctx, cancelFunc := context.WithCancel(ctx)
	// LATER: handle jobregister error
	if err := srvutils.JobRegister(ctx, cancelFunc, "Delete network "+in.GetName()); err == nil {
		defer srvutils.JobDeregister(ctx)
	} /* else {
		return empty, scerr.InvalidInstanceContentError("ctx", "has no uuid").ToGRPCStatus()
	}*/

	tenant := GetCurrentTenant()
	if tenant == nil {
		msg := "cannot delete network: no tenant set"
		tracer.Trace(utils.Capitalize(msg))
		return empty, status.Errorf(codes.FailedPrecondition, msg)
	}

	handler := handlers.NewNetworkHandler(currentTenant.Service)
	err = handler.Delete(ctx, ref)
	if err != nil {
		return empty, scerr.Wrap(err, "cannot delete network").ToGRPCStatus()
	}

	tracer.Trace("Network '%s' successfully deleted.", ref)
	return empty, nil
}
