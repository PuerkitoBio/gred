package cmd

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/PuerkitoBio/gred/srv"
)

const (
	WrongNumberOfArgsFmt = "ERR wrong number of arguments for '%s' command"
)

var (
	ErrArgNotInteger  = errors.New("ERR value is not an integer or out of range")
	ErrArgNotFloat    = errors.New("ERR value is not a valid float")
	ErrInvalidValType = errors.New("ERR Operation against a key holding the wrong kind of value")
	ErrNilSuccess     = errors.New("nil")
	ErrPong           = errors.New("pong")
)

var Commands = make(map[string]Cmd)

func Register(name string, c Cmd) {
	if name == "" {
		panic("cmds: call Register with empty command name")
	}
	if _, ok := Commands[name]; ok {
		panic(fmt.Sprintf("cmds: command %s already registered", name))
	}
	Commands[name] = c
}

type Cmd interface {
	GetArgDef() *ArgDef
}

type DBCmd interface {
	Cmd
	ExecWithDB(srv.DB, []string, []int64, []float64) (interface{}, error)
}

type ArgFn func([]string, []int64, []float64) error

type ArgDef struct {
	FloatIndices     []int
	IntIndices       []int
	MinArgs, MaxArgs int
	ValidateFn       ArgFn
}

func (a *ArgDef) GetArgDef() *ArgDef { return a }

func (a *ArgDef) ParseArgs(name string, args []string) ([]string, []int64, []float64, error) {
	l := len(args)
	if l < a.MinArgs || (l > a.MaxArgs && a.MaxArgs >= 0) {
		return nil, nil, nil, fmt.Errorf(WrongNumberOfArgsFmt, name)
	}

	// Parse integers
	intix := a.IntIndices
	ints := make([]int64, len(intix))
	for i, ix := range intix {
		if ix < 0 {
			ix = l + ix
		}
		val, err := strconv.ParseInt(args[ix], 10, 64)
		if err != nil {
			return nil, nil, nil, ErrArgNotInteger
		}
		ints[i] = val
	}

	// Parse floats
	fix := a.FloatIndices
	floats := make([]float64, len(fix))
	for i, ix := range fix {
		if ix < 0 {
			ix = l + ix
		}
		val, err := strconv.ParseFloat(args[ix], 64)
		if err != nil {
			return nil, nil, nil, ErrArgNotFloat
		}
		floats[i] = val
	}

	if a.ValidateFn != nil {
		err := a.ValidateFn(args, ints, floats)
		if err != nil {
			return nil, nil, nil, err
		}
	}
	return args, ints, floats, nil
}

var _ DBCmd = (*singleKeyCmd)(nil)

type KeyFn func(srv.Key, []string, []int64, []float64) (interface{}, error)

func NewSingleKeyCmd(arg *ArgDef, noKeyFlag srv.NoKeyFlag, fn KeyFn) DBCmd {
	return &singleKeyCmd{
		ArgDef: arg,
		noKey:  noKeyFlag,
		fn:     fn,
	}
}

type singleKeyCmd struct {
	*ArgDef
	noKey srv.NoKeyFlag
	fn    KeyFn
}

func (c *singleKeyCmd) ExecWithDB(db srv.DB, args []string, ints []int64, floats []float64) (interface{}, error) {
	k, def := db.LockGetKey(args[0], c.noKey)
	defer def()

	return c.fn(k, args, ints, floats)
}

var _ DBCmd = (*dbCmd)(nil)

type DBFn func(srv.DB, []string, []int64, []float64) (interface{}, error)

func NewDBCmd(arg *ArgDef, fn DBFn) DBCmd {
	return &dbCmd{
		ArgDef: arg,
		fn:     fn,
	}
}

type dbCmd struct {
	*ArgDef
	fn DBFn
}

func (d *dbCmd) ExecWithDB(db srv.DB, args []string, ints []int64, floats []float64) (interface{}, error) {
	return d.fn(db, args, ints, floats)
}