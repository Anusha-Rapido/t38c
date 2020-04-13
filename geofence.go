package t38c

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
)

// GeofenceRequestable interface
// TODO: rename?
type GeofenceRequestable interface {
	GeofenceCommand() Command
}

var _ GeofenceRequestable = (*Request)(nil)

// Request struct
type Request struct {
	Cmd           string
	Key           string
	Area          Command
	OutputFormat  OutputFormat
	DetectActions []DetectAction
	Options       []SearchOption
}

// GeofenceCommand ...
func (req *Request) GeofenceCommand() Command {
	var args []string
	args = append(args, req.Key)

	for _, opt := range req.Options {
		args = append(args, opt.Name)
		args = append(args, opt.Args...)
	}

	args = append(args, "FENCE")

	if len(req.DetectActions) > 0 {
		args = append(args, "DETECT")
		actions := ""
		first := true
		for _, action := range req.DetectActions {
			if !first {
				actions += ","
			}
			actions += string(action)
			first = false
		}
		args = append(args, actions)
	}

	if len(req.OutputFormat.Name) > 0 {
		args = append(args, req.OutputFormat.Name)
		args = append(args, req.OutputFormat.Args...)
	}

	args = append(args, req.Area.Name)
	args = append(args, req.Area.Args...)

	return NewCommand(req.Cmd, args...)
}

// Actions ...
func (req *Request) Actions(actions ...DetectAction) *Request {
	req.DetectActions = actions
	return req
}

// WithOptions ...
func (req *Request) WithOptions(opts ...SearchOption) *Request {
	req.Options = opts
	return req
}

// Format ...
func (req *Request) Format(fmt OutputFormat) *Request {
	req.OutputFormat = fmt
	return req
}

// GeofenceWithin ...
func GeofenceWithin(key string, area SearchArea) *Request {
	return &Request{
		Cmd:  "WITHIN",
		Key:  key,
		Area: Command(area),
	}
}

// GeofenceIntersects ...
func GeofenceIntersects(key string, area SearchArea) *Request {
	return &Request{
		Cmd:  "INTERSECTS",
		Key:  key,
		Area: Command(area),
	}
}

// GeofenceNearby ...
func GeofenceNearby(key string, lat, lon, meters float64) *Request {
	return &Request{
		Cmd:  "NEARBY",
		Key:  key,
		Area: NewCommand("POINT", floatString(lat), floatString(lon), floatString(meters)),
	}
}

var _ GeofenceRequestable = (*RoamRequest)(nil)

// RoamRequest struct
type RoamRequest struct {
	Key           string
	Target        string
	Pattern       string
	Meters        int
	OutputFormat  OutputFormat
	DetectActions []DetectAction
	Options       []SearchOption
}

// GeofenceCommand ...
func (req *RoamRequest) GeofenceCommand() Command {
	var args []string
	args = append(args, req.Key)

	for _, opt := range req.Options {
		args = append(args, opt.Name)
		args = append(args, opt.Args...)
	}

	args = append(args, "FENCE")

	if len(req.DetectActions) > 0 {
		args = append(args, "DETECT")
		actions := ""
		first := true
		for _, action := range req.DetectActions {
			if !first {
				actions += ","
			}
			actions += string(action)
			first = false
		}
		args = append(args, actions)
	}

	if len(req.OutputFormat.Name) > 0 {
		args = append(args, req.OutputFormat.Name)
		args = append(args, req.OutputFormat.Args...)
	}

	args = append(args, []string{
		"ROAM", req.Target, req.Pattern, strconv.Itoa(req.Meters),
	}...)

	return NewCommand("NEARBY", args...)
}

// Actions ...
func (req *RoamRequest) Actions(actions ...DetectAction) *RoamRequest {
	req.DetectActions = actions
	return req
}

// WithOptions ...
func (req *RoamRequest) WithOptions(opts ...SearchOption) *RoamRequest {
	req.Options = opts
	return req
}

// Format ...
func (req *RoamRequest) Format(fmt OutputFormat) *RoamRequest {
	req.OutputFormat = fmt
	return req
}

// GeofenceRoam ...
func GeofenceRoam(key, target, pattern string, meters int) *RoamRequest {
	return &RoamRequest{
		Key:     key,
		Target:  target,
		Pattern: pattern,
		Meters:  meters,
	}
}

// Fence ...
func (client *Client) Fence(ctx context.Context, req GeofenceRequestable) (chan Response, error) {
	cmd := req.GeofenceCommand()
	events, err := client.ExecuteStream(ctx, cmd.Name, cmd.Args...)
	if err != nil {
		return nil, err
	}

	ch := make(chan Response, 10)
	go func() {
		defer close(ch)
		for event := range events {
			var resp Response
			if err := json.Unmarshal(event, &resp); err != nil {
				log.Printf("bad event: %v", err)
				break
			}

			ch <- resp
		}
	}()

	return ch, nil
}