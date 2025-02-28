package daemon

import (
	"fmt"
	"gearbox/eventbroker/eblog"
	"gearbox/eventbroker/entity"
	"gearbox/eventbroker/msgs"
	"gearbox/eventbroker/osdirs"
	"gearbox/eventbroker/states"
	"gearbox/eventbroker/tasks"
	"github.com/gearboxworks/go-status/only"
	"time"
)

func (me *Daemon) New(args ...Args) error {

	var _args Args
	var err error

	for range only.Once {

		if len(args) > 0 {
			_args = args[0]
		}

		if _args.Channels == nil {
			err = msgs.MakeError(me.EntityId, "channel pointer is nil")
			break
		}

		if _args.BaseDirs == nil {
			err = msgs.MakeError(me.EntityId, "ospaths is nil")
			break
		}

		if _args.EntityId == "" {
			_args.EntityId = entity.DaemonEntityName
		}
		_args.State = states.New(_args.EntityId, _args.EntityId, entity.SelfEntityName)

		if _args.Boxname == "" {
			_args.Boxname = entity.DaemonEntityName
		}

		jdir := osdirs.AddPaths(_args.BaseDirs.EventBrokerEtcDir, DefaultJsonDir)
		_, err = osdirs.CreateIfNotExists(jdir)
		if err != nil {
			break
		}

		_args.daemons = make(ServicesMap) // Mutex not required

		_args.State.SetWant(states.StateIdle)
		_args.State.SetNewState(states.StateIdle, err)

		*me = Daemon(_args)
		eblog.Debug(me.EntityId, "init complete")
	}

	me.Channels.PublishState(me.State)
	eblog.LogIfNil(me, err)
	eblog.LogIfError(err)

	return err
}

// Start the M-DNS network handler.
func (me *Daemon) StartHandler() error {

	var err error

	for range only.Once {
		err = me.EnsureNotNil()
		if err != nil {
			break
		}

		me.State.SetNewAction(states.ActionStart)
		me.Channels.PublishState(me.State)

		for range only.Once {
			me.Task, err = tasks.StartTask(initDaemon, startDaemon, monitorDaemon, stopDaemon, me)
			if err != nil {
				break
			}
		}

		me.State.SetNewState(states.StateStarted, err)
		me.Channels.PublishState(me.State)
		eblog.Debug(me.EntityId, "started task handler")
	}

	eblog.LogIfNil(me, err)
	eblog.LogIfError(err)

	return err
}

// Stop the daemon handler.
func (me *Daemon) StopHandler() error {

	var err error

	for range only.Once {
		err = me.EnsureNotNil()
		if err != nil {
			break
		}

		me.State.SetNewAction(states.ActionStop)
		me.Channels.PublishState(me.State)

		for range only.Once {
			_ = me.StopServices()
			// Ignore error, will clean up when program exits.

			err = me.Task.Stop()
		}

		me.State.SetNewState(states.StateStopped, err)
		me.Channels.PublishState(me.State)
		eblog.Debug(me.EntityId, "stopped task handler")
	}

	eblog.LogIfNil(me, err)
	eblog.LogIfError(err)

	return err
}

func (me *Daemon) StopServices() error {

	var err error

	for range only.Once {
		err = me.EnsureNotNil()
		if err != nil {
			break
		}

		for _, u := range me.GetManagedEntities() { // Ignore Mutex
			err = me.daemons[u].Stop() // Ignore Mutex
			if err == nil {
				err = me.DeleteEntity(u)
			}
			// Ignore error, will clean up when program exits.
		}
	}

	eblog.LogIfNil(me, err)
	eblog.LogIfError(err)

	return err
}

// Print all services registered under daemon that I manage.
func (me *Daemon) PrintServices() error {

	var err error

	for range only.Once {
		err = me.EnsureNotNil()
		if err != nil {
			break
		}

		_ = me.daemons.Print()
	}

	return err
}

// Print all services registered under daemon that I manage.
func (me *Daemon) ListStarted() (msgs.Addresses, error) {

	var err error
	var sc msgs.Addresses

	for range only.Once {
		err = me.EnsureNotNil()
		if err != nil {
			break
		}

		for _, u := range me.GetManagedEntities() {
			fmt.Printf("# Entry: %s\n", u)
			sc = append(sc, u)
		}

		_ = me.daemons.Print()
	}

	return sc, err
}

func (me *Daemon) TestMe() error {

	var err error

	fmt.Printf("DEBUG STARTED\n")

	var s *Service
	s, err = me.RegisterByFile("/Users/mick/.gearbox/dist/eventbroker/unfsd/unfsd.json")

	time.Sleep(time.Second * 8)
	if err == nil {
		var state states.Status

		state, err = s.Status(PublishState)
		fmt.Printf("Status: %v\n", state)
		if err != nil {
			fmt.Printf("Woops!\n")
		}

		//
		//err = s.instance.service.Start()
		//if err != nil {
		//	fmt.Printf("Woops!\n")
		//}
		//
		//state, err = s.Status()
		//fmt.Printf("Status: %v\n", state)
		//if err != nil {
		//	fmt.Printf("Woops!\n")
		//}
		//
		//err = s.Stop()
		//if err != nil {
		//	fmt.Printf("Woops!\n")
		//}
		//
		//err = s.instance.service.Uninstall()
		//if err != nil {
		//	fmt.Printf("Woops!\n")
		//}
	}

	fmt.Printf("DEBUG SLEEPING")
	time.Sleep(time.Hour * 4200)

	eblog.LogIfNil(me, err)
	eblog.LogIfError(err)

	return err
}
