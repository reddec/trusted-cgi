// Code generated by jsonrpc2. DO NOT EDIT.
package handlers

import (
	"context"
	"encoding/json"
	jsonrpc2 "github.com/reddec/jsonrpc2"
	api "github.com/reddec/trusted-cgi/api"
	application "github.com/reddec/trusted-cgi/application"
)

func RegisterQueuesAPI(router *jsonrpc2.Router, wrap api.QueuesAPI, typeHandler interface {
	ValidateToken(ctx context.Context, value *api.Token) error
}) []string {
	router.RegisterFunc("QueuesAPI.Create", func(ctx context.Context, params json.RawMessage, positional bool) (interface{}, error) {
		var args struct {
			Arg0 *api.Token        `json:"token"`
			Arg1 application.Queue `json:"queue"`
		}
		var err error
		if positional {
			err = jsonrpc2.UnmarshalArray(params, &args.Arg0, &args.Arg1)
		} else {
			err = json.Unmarshal(params, &args)
		}
		if err != nil {
			return nil, err
		}
		err = typeHandler.ValidateToken(ctx, args.Arg0)
		if err != nil {
			return nil, err
		}
		return wrap.Create(ctx, args.Arg0, args.Arg1)
	})

	router.RegisterFunc("QueuesAPI.Remove", func(ctx context.Context, params json.RawMessage, positional bool) (interface{}, error) {
		var args struct {
			Arg0 *api.Token `json:"token"`
			Arg1 string     `json:"name"`
		}
		var err error
		if positional {
			err = jsonrpc2.UnmarshalArray(params, &args.Arg0, &args.Arg1)
		} else {
			err = json.Unmarshal(params, &args)
		}
		if err != nil {
			return nil, err
		}
		err = typeHandler.ValidateToken(ctx, args.Arg0)
		if err != nil {
			return nil, err
		}
		return wrap.Remove(ctx, args.Arg0, args.Arg1)
	})

	router.RegisterFunc("QueuesAPI.Linked", func(ctx context.Context, params json.RawMessage, positional bool) (interface{}, error) {
		var args struct {
			Arg0 *api.Token `json:"token"`
			Arg1 string     `json:"lambda"`
		}
		var err error
		if positional {
			err = jsonrpc2.UnmarshalArray(params, &args.Arg0, &args.Arg1)
		} else {
			err = json.Unmarshal(params, &args)
		}
		if err != nil {
			return nil, err
		}
		err = typeHandler.ValidateToken(ctx, args.Arg0)
		if err != nil {
			return nil, err
		}
		return wrap.Linked(ctx, args.Arg0, args.Arg1)
	})

	router.RegisterFunc("QueuesAPI.List", func(ctx context.Context, params json.RawMessage, positional bool) (interface{}, error) {
		var args struct {
			Arg0 *api.Token `json:"token"`
		}
		var err error
		if positional {
			err = jsonrpc2.UnmarshalArray(params, &args.Arg0)
		} else {
			err = json.Unmarshal(params, &args)
		}
		if err != nil {
			return nil, err
		}
		err = typeHandler.ValidateToken(ctx, args.Arg0)
		if err != nil {
			return nil, err
		}
		return wrap.List(ctx, args.Arg0)
	})

	router.RegisterFunc("QueuesAPI.Assign", func(ctx context.Context, params json.RawMessage, positional bool) (interface{}, error) {
		var args struct {
			Arg0 *api.Token `json:"token"`
			Arg1 string     `json:"name"`
			Arg2 string     `json:"lambda"`
		}
		var err error
		if positional {
			err = jsonrpc2.UnmarshalArray(params, &args.Arg0, &args.Arg1, &args.Arg2)
		} else {
			err = json.Unmarshal(params, &args)
		}
		if err != nil {
			return nil, err
		}
		err = typeHandler.ValidateToken(ctx, args.Arg0)
		if err != nil {
			return nil, err
		}
		return wrap.Assign(ctx, args.Arg0, args.Arg1, args.Arg2)
	})

	return []string{"QueuesAPI.Create", "QueuesAPI.Remove", "QueuesAPI.Linked", "QueuesAPI.List", "QueuesAPI.Assign"}
}
