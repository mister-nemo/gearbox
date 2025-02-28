package channels

import (
	"gearbox/eventbroker/eblog"
	"gearbox/eventbroker/entity"
	"gearbox/eventbroker/msgs"
	"gearbox/eventbroker/states"
	"github.com/gearboxworks/go-status/only"
	"time"
)

func (me *Channels) Publish(msg msgs.Message) error {

	var err error

	for range only.Once {
		err = me.EnsureNotNil()
		if err != nil {
			break
		}

		err = msg.Topic.EnsureNotNil()
		if err != nil {
			break
		}

		if msg.Time.IsNil() {
			msg.Time = msg.Time.Now()
		}

		if msg.Source.EnsureNotEmpty() != nil {
			msg.Source = me.EntityId
		}

		if msg.Topic.Address.EnsureNotEmpty() != nil {
			err = msgs.MakeError(me.EntityId, "no destination for channel message")
			break
		}

		// eblog.Debug(me.EntityId, "Publish(%s) =>\tmsg.NewTopic():%v\tme.instance.emitter:%v", msg.Topic.String(), msg, me.instance.emitter)
		me.instance.emits = me.instance.emitter.Emit(msg.Topic.String(), msg)
		if me.instance.emits == nil {
			err = msgs.MakeError(me.EntityId, "failed to send channel message")
			break
		}

		eblog.Debug(me.EntityId, "Channel time:%d src:%s topic:%s msg:%s", msg.Time.Unix(), msg.Source.String(), msg.Topic.String(), msg.Text.String())
		/*
			select {
				case <-me.emits:
					// err = msgs.MakeError(me.EntityId,"channel message sent OK")

				case <-time.After(time.Second * 10):
					err = msgs.MakeError(me.EntityId,"timeout sending channel message")
					close(me.emits)
			}
		*/
	}

	eblog.LogIfNil(me, err)
	eblog.LogIfError(err)

	return err
}

func (me *Channels) GetCallbackReturn(msg msgs.Message, waitForExecute int) (Return, error) {

	var ret Return
	var err error

	for range only.Once {
		err = me.EnsureNotNil()
		if err != nil {
			break
		}

		client := msg.Topic.Address
		if _, ok := me.subscribers[client]; !ok {
			err = msgs.MakeError(me.EntityId, "unknown channel subscriber")
			break
		}

		subtopic := msg.Topic.SubTopic
		// MUTEX if _, ok := me.subscribers[client].Returns[subtopic]; !ok {
		err, _, _, _, _ = me.subscribers[client].GetTopic(subtopic)
		if err != nil {
			break
		}

		// MUTEX for loop := 0; (me.subscribers[client].Executed[subtopic] == false) && (loop < waitForExecute); loop++ {
		for loop := 0; (me.subscribers[client].GetExecuted(subtopic) == false) && (loop < waitForExecute); loop++ {
			// Wait if we are asked to.
			time.Sleep(time.Millisecond * 10)
		}

		// MUTEX if me.subscribers[client].Executed[subtopic] == false {
		if me.subscribers[client].GetExecuted(subtopic) == false {
			err = msgs.MakeError(me.EntityId, "no response from channel")
			break
		}

		// MUTEX ret = me.subscribers[client].Returns[subtopic]
		ret = me.subscribers[client].GetReturns(subtopic)
	}

	eblog.LogIfNil(me, err)
	eblog.LogIfError(err)

	return ret, err
}

func (me *Channels) SetCallbackReturnToNil(msg msgs.Message) error {

	var err error

	for range only.Once {
		err = me.EnsureNotNil()
		if err != nil {
			break
		}

		err = me.EnsureSubscriberNotNil(msg.Topic.Address)
		if err != nil {
			break
		}

		err, _, _, _, _ = me.subscribers[msg.Topic.Address].GetTopic(msg.Topic.SubTopic)
		if err != nil {
			break
		}

		// MUTEX me.subscribers[client].Returns[subtopic] = nil
		me.subscribers[msg.Topic.Address].SetReturns(msg.Topic.SubTopic, nil)
	}

	eblog.LogIfNil(me, err)
	eblog.LogIfError(err)

	return err
}

func (me *Channels) PublishAndWaitForReturn(msg msgs.Message, waitForExecute int) (Return, error) {

	var err error
	var ret Return

	for range only.Once {
		err = me.EnsureNotNil()
		if err != nil {
			break
		}

		err = me.Publish(msg)
		if err != nil {
			break
		}

		ret, err = me.GetCallbackReturn(msg, waitForExecute)
		if err != nil {
			break
		}

		eblog.Debug(me.EntityId, "message returned by channel OK")
	}

	eblog.LogIfNil(me, err)
	eblog.LogIfError(err)

	return ret, err
}

// Send channel message on state changes only.
//func PublishState(me *Channels, caller *msg.Address, state *states.Status) {
func PublishState(me *Channels, state *states.Status) {

	switch {
	case me == nil:
		//fmt.Printf("me == nil: %s\n", state.String())
	case state == nil:
		//fmt.Printf("state == nil: %s\n", state.String())
	case state.EnsureNotNil() != nil:
		//fmt.Printf("state.EnsureNotEmpty() != nil: %s\n", state.String())
	case !state.HasChangedState():
		//fmt.Printf("!state.HasChangedState(): %s\n", state.String())

	case state.GetError() != nil:
		//msg := state.EntityId.MakeMessage(entity.BroadcastEntityName, states.ActionError, msg.Text(state.GetError().Error()))
		msg := state.EntityId.MakeMessage(entity.BroadcastEntityName, states.ActionError, state.ToMessageText())
		//fmt.Printf("ERROR: %s\n", msg.String())
		_ = me.Publish(msg)

	case state.ExpectingNewState():
		fallthrough
	case state.HasChangedState():
		//msg := state.EntityId.MakeMessage(entity.BroadcastEntityName, states.ActionStatus, msg.Text(state.GetCurrent()))
		msg := state.EntityId.MakeMessage(entity.BroadcastEntityName, states.ActionStatus, state.ToMessageText())
		//fmt.Printf("EXPECTING: %s\n", msg.String())
		_ = me.Publish(msg)
	}

	return
}

func (me *Channels) PublishState(state *states.Status) {

	PublishState(me, state)

	return
}

func (me *Channels) PublishSpecificState(caller *msgs.Address, state states.State) {

	switch {
	case me == nil:
	case state == "":
	case caller == nil:
		return
	}

	msg := caller.MakeMessage(entity.BroadcastEntityName, states.ActionStatus, msgs.Text(state))
	_ = me.Publish(msg)

	return
}
