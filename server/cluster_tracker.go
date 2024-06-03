package server

import (
	"context"
	syncAtomic "sync/atomic"

	ncapi "github.com/doublemo/nakama-cluster/api"
	"github.com/gofrs/uuid"
	"github.com/heroiclabs/nakama-common/runtime"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

func (s *ClusterServer) NotifyTrack(presences ...*Presence) error {
	track := &ncapi.Message_Track{Presences: make([]*ncapi.Message_Presence, len(presences))}

	for i, presence := range presences {
		track.Presences[i] = &ncapi.Message_Presence{
			Id: &ncapi.Message_PresenceID{
				Node:      presence.GetNodeId(),
				SessionID: presence.GetSessionId(),
			},
			Stream: &ncapi.Message_PresenceStream{
				Mode:       int32(presence.Stream.Mode),
				Subject:    presence.Stream.Subject.String(),
				Subcontext: presence.Stream.Subcontext.String(),
				Label:      presence.Stream.Label,
			},

			UserID: presence.GetUserId(),
			Meta: &ncapi.Message_PresenceMeta{
				Format:      int32(presence.Meta.Format),
				Hidden:      presence.Meta.Hidden,
				Persistence: presence.Meta.Persistence,
				Username:    presence.Meta.Username,
				Status:      presence.Meta.Status,
				Reason:      uint32(presence.Meta.Reason),
			},
		}
	}

	// Serialize Track message to bytes
	payloadBytes, _ := proto.Marshal(track)

	return s.Broadcast(&ncapi.Envelope{
		Payload: &ncapi.Envelope_Bytes{Bytes: payloadBytes},
		Cid:     ncapi.Message_TRACK,
	})
}

func (s *ClusterServer) NotifyUntrack(presences ...*Presence) error {
	untrack := &ncapi.Message_Untrack{Presences: make([]*ncapi.Message_Presence, len(presences))}

	for i, presence := range presences {
		untrack.Presences[i] = &ncapi.Message_Presence{
			Id: &ncapi.Message_PresenceID{
				Node:      presence.GetNodeId(),
				SessionID: presence.GetSessionId(),
			},
			Stream: &ncapi.Message_PresenceStream{
				Mode:       int32(presence.Stream.Mode),
				Subject:    presence.Stream.Subject.String(),
				Subcontext: presence.Stream.Subcontext.String(),
				Label:      presence.Stream.Label,
			},

			UserID: presence.GetUserId(),
			Meta: &ncapi.Message_PresenceMeta{
				Format:      int32(presence.Meta.Format),
				Hidden:      presence.Meta.Hidden,
				Persistence: presence.Meta.Persistence,
				Username:    presence.Meta.Username,
				Status:      presence.Meta.Status,
				Reason:      uint32(presence.Meta.Reason),
			},
		}
	}
	payloadBytes, _ := proto.Marshal(untrack)

	return s.Broadcast(&ncapi.Envelope{
		Payload: &ncapi.Envelope_Bytes{Bytes: payloadBytes},
		Cid:     ncapi.Message_UNTRACK,
	})
}

func (s *ClusterServer) NotifyUntrackAll(sessionID uuid.UUID, reason runtime.PresenceReason) error {
	//untrackAll := ncapi.Message_UntrackAll{SessionID: sessionID.String(), Reason: int32(reason)}

	// return s.Broadcast(&ncapi.Envelope{Payload: &ncapi.Message_UntrackAll{UntrackAll: &untrackAll}})
	return nil
}

func (s *ClusterServer) NotifyUntrackByMode(sessionID uuid.UUID, modes map[uint8]struct{}, skipStream PresenceStream) error {
	untrack := ncapi.Message_UntrackByMode{
		SessionID: sessionID.String(),
		Modes:     make([]int32, len(modes)),
		SkipStream: &ncapi.Message_PresenceStream{
			Mode:       int32(skipStream.Mode),
			Subject:    skipStream.Subject.String(),
			Subcontext: skipStream.Subcontext.String(),
			Label:      skipStream.Label,
		},
	}

	i := 0
	for m := range modes {
		untrack.Modes[i] = int32(m)
		i++
	}

	//return s.Broadcast(&ncapi.Message_Envelope{Payload: &ncapi.Message_Envelope_UntrackByMode{UntrackByMode: &untrack}})
	return nil
}

func (s *ClusterServer) NotifyUntrackByStream(streams ...PresenceStream) error {
	untrack := ncapi.Message_UntrackByStream{Streams: make([]*ncapi.Message_PresenceStream, len(streams))}

	for i, stream := range streams {
		untrack.Streams[i] = &ncapi.Message_PresenceStream{
			Mode:       int32(stream.Mode),
			Subject:    stream.Subject.String(),
			Subcontext: stream.Subcontext.String(),
			Label:      stream.Label,
		}
	}

	//return s.Broadcast(&ncapi.Message_Envelope{Payload: &ncapi.Message_Envelope_UntrackByStream{UntrackByStream: &untrack}})
	return nil
}

func (s *ClusterServer) onTrack(node string, msg ncapi.Message_Track) {
	s.logger.Debug("onTrack", zap.String("node", node))
	for _, presence := range msg.GetPresences() {
		if presence.Meta.Reason == uint32(runtime.PresenceReasonUpdate) {
			s.tracker.UpdateFromNode(s.ctx,
				presence.Id.Node,
				uuid.FromStringOrNil(presence.Id.SessionID),
				PresenceStream{
					Mode:       uint8(presence.Stream.Mode),
					Subject:    uuid.FromStringOrNil(presence.Stream.Subject),
					Subcontext: uuid.FromStringOrNil(presence.Stream.Subcontext),
					Label:      presence.Stream.Label,
				},
				uuid.FromStringOrNil(presence.UserID),
				PresenceMeta{
					Format:      SessionFormat(presence.Meta.Format),
					Hidden:      presence.Meta.Hidden,
					Persistence: presence.Meta.Persistence,
					Username:    presence.Meta.Username,
					Status:      presence.Meta.Status,
					Reason:      uint32(presence.Meta.Reason),
				}, true)
			continue
		}

		s.tracker.TrackFromNode(s.ctx,
			presence.Id.Node,
			uuid.FromStringOrNil(presence.Id.SessionID),
			PresenceStream{
				Mode:       uint8(presence.Stream.Mode),
				Subject:    uuid.FromStringOrNil(presence.Stream.Subject),
				Subcontext: uuid.FromStringOrNil(presence.Stream.Subcontext),
				Label:      presence.Stream.Label,
			},
			uuid.FromStringOrNil(presence.UserID),
			PresenceMeta{
				Format:      SessionFormat(presence.Meta.Format),
				Hidden:      presence.Meta.Hidden,
				Persistence: presence.Meta.Persistence,
				Username:    presence.Meta.Username,
				Status:      presence.Meta.Status,
				Reason:      uint32(presence.Meta.Reason),
			}, true)
	}
}

func (s *ClusterServer) onUntrack(node string, msg ncapi.Message_Untrack) {
	s.logger.Debug("onUntrack", zap.String("node", node))

	for _, presence := range msg.Presences {
		s.tracker.UntrackFromNode(presence.Id.Node, uuid.FromStringOrNil(presence.Id.SessionID),
			PresenceStream{
				Mode:       uint8(presence.Stream.Mode),
				Subject:    uuid.FromStringOrNil(presence.Stream.Subject),
				Subcontext: uuid.FromStringOrNil(presence.Stream.Subcontext),
				Label:      presence.Stream.Label,
			},
			uuid.FromStringOrNil(presence.UserID))
	}
}

func (s *ClusterServer) onUntrackAll(node string, msg *ncapi.Message_Envelope) {
	s.logger.Debug("onUntrackAll", zap.String("node", node))
	var message ncapi.Message_UntrackAll
	if err := proto.Unmarshal(msg.GetMessage(), &message); err != nil {
		s.logger.Warn("NotifyMsg parse failed", zap.Error(err))
		return
	}

	s.tracker.UntrackAllFromNode(node, uuid.FromStringOrNil(message.SessionID), runtime.PresenceReason(message.Reason))
}

func (s *ClusterServer) onUntrackByMode(node string, msg *ncapi.Message_Envelope) {
	s.logger.Debug("onUntrackByMode", zap.String("node", node))
	var message ncapi.Message_UntrackByMode
	if err := proto.Unmarshal(msg.GetMessage(), &message); err != nil {
		s.logger.Warn("NotifyMsg parse failed", zap.Error(err))
		return
	}

	modes := make(map[uint8]struct{})
	for _, mode := range message.Modes {
		modes[uint8(mode)] = struct{}{}
	}

	s.tracker.UntrackByModesFromNode(node, uuid.FromStringOrNil(message.SessionID), modes, PresenceStream{
		Mode:       uint8(message.SkipStream.Mode),
		Subject:    uuid.FromStringOrNil(message.SkipStream.Subject),
		Subcontext: uuid.FromStringOrNil(message.SkipStream.Subcontext),
		Label:      message.SkipStream.Label,
	})
}

func (s *ClusterServer) onUntrackByStream(node string, msg *ncapi.Message_Envelope) {
	s.logger.Debug("onUntrackByStream", zap.String("node", node))
	var message ncapi.Message_UntrackByStream
	if err := proto.Unmarshal(msg.GetMessage(), &message); err != nil {
		s.logger.Warn("NotifyMsg parse failed", zap.Error(err))
		return
	}

	for _, stream := range message.Streams {
		s.tracker.UntrackByStreamFromNode(node, PresenceStream{
			Mode:       uint8(stream.Mode),
			Subject:    uuid.FromStringOrNil(stream.Subject),
			Subcontext: uuid.FromStringOrNil(stream.Subcontext),
			Label:      stream.Label,
		})
	}
}

func (t *LocalTracker) TrackFromNode(ctx context.Context, node string, sessionID uuid.UUID, stream PresenceStream, userID uuid.UUID, meta PresenceMeta, allowIfFirstForSession bool) (bool, bool) {
	syncAtomic.StoreUint32(&meta.Reason, uint32(runtime.PresenceReasonJoin))
	pc := presenceCompact{ID: PresenceID{Node: node, SessionID: sessionID}, Stream: stream, UserID: userID}
	p := &Presence{ID: PresenceID{Node: node, SessionID: sessionID}, Stream: stream, UserID: userID, Meta: meta}
	t.Lock()

	select {
	case <-ctx.Done():
		t.Unlock()
		return false, false
	default:
	}

	// See if this session has any presences tracked at all.
	if bySession, anyTracked := t.presencesBySession[sessionID]; anyTracked {
		// Then see if the exact presence we need is tracked.
		if _, alreadyTracked := bySession[pc]; !alreadyTracked {
			// If the current session had others tracked, but not this presence.
			bySession[pc] = p
		} else {
			t.Unlock()
			return true, false
		}
	} else {
		if !allowIfFirstForSession {
			// If it's the first presence for this session, only allow it if explicitly permitted to.
			t.Unlock()
			return false, false
		}
		// If nothing at all was tracked for the current session, begin tracking.
		bySession = make(map[presenceCompact]*Presence)
		bySession[pc] = p
		t.presencesBySession[sessionID] = bySession
	}
	t.count.Inc()

	// Update tracking for stream.
	byStreamMode, ok := t.presencesByStream[stream.Mode]
	if !ok {
		byStreamMode = make(map[PresenceStream]map[presenceCompact]*Presence)
		t.presencesByStream[stream.Mode] = byStreamMode
	}

	if byStream, ok := byStreamMode[stream]; !ok {
		byStream = make(map[presenceCompact]*Presence)
		byStream[pc] = p
		byStreamMode[stream] = byStream
	} else {
		byStream[pc] = p
	}

	t.Unlock()
	if !meta.Hidden {
		t.queueEvent([]*Presence{p}, nil)
	}
	return true, true
}

func (t *LocalTracker) UntrackFromNode(node string, sessionID uuid.UUID, stream PresenceStream, userID uuid.UUID) {
	pc := presenceCompact{ID: PresenceID{Node: node, SessionID: sessionID}, Stream: stream, UserID: userID}
	t.Lock()

	bySession, anyTracked := t.presencesBySession[sessionID]
	if !anyTracked {
		// Nothing tracked for the session.
		t.Unlock()
		return
	}
	p, found := bySession[pc]
	if !found {
		// The session had other presences, but not for this stream.
		t.Unlock()
		return
	}

	// Update the tracking for session.
	if len(bySession) == 1 {
		// This was the only presence for the session, discard the whole list.
		delete(t.presencesBySession, sessionID)
	} else {
		// There were other presences for the session, drop just this one.
		delete(bySession, pc)
	}
	t.count.Dec()

	// Update the tracking for stream.
	if byStreamMode := t.presencesByStream[stream.Mode]; len(byStreamMode) == 1 {
		// This is the only stream for this stream mode.
		if byStream := byStreamMode[stream]; len(byStream) == 1 {
			// This was the only presence in the only stream for this stream mode, discard the whole list.
			delete(t.presencesByStream, stream.Mode)
		} else {
			// There were other presences for the stream, drop just this one.
			delete(byStream, pc)
		}
	} else {
		// There are other streams for this stream mode.
		if byStream := byStreamMode[stream]; len(byStream) == 1 {
			// This was the only presence for the stream, discard the whole list.
			delete(byStreamMode, stream)
		} else {
			// There were other presences for the stream, drop just this one.
			delete(byStream, pc)
		}
	}

	t.Unlock()
	if !p.Meta.Hidden {
		syncAtomic.StoreUint32(&p.Meta.Reason, uint32(runtime.PresenceReasonLeave))
		t.queueEvent(nil, []*Presence{p})
	}
}

func (t *LocalTracker) UpdateFromNode(ctx context.Context, node string, sessionID uuid.UUID, stream PresenceStream, userID uuid.UUID, meta PresenceMeta, allowIfFirstForSession bool) bool {
	syncAtomic.StoreUint32(&meta.Reason, uint32(runtime.PresenceReasonUpdate))
	pc := presenceCompact{ID: PresenceID{Node: node, SessionID: sessionID}, Stream: stream, UserID: userID}
	p := &Presence{ID: PresenceID{Node: node, SessionID: sessionID}, Stream: stream, UserID: userID, Meta: meta}
	t.Lock()

	select {
	case <-ctx.Done():
		t.Unlock()
		return false
	default:
	}

	bySession, anyTracked := t.presencesBySession[sessionID]
	if !anyTracked {
		if !allowIfFirstForSession {
			// Nothing tracked for the session and not allowed to track as first presence.
			t.Unlock()
			return false
		}

		bySession = make(map[presenceCompact]*Presence)
		t.presencesBySession[sessionID] = bySession
	}

	// Update tracking for session, but capture any previous meta in case a leave event is required.
	previousP, alreadyTracked := bySession[pc]
	bySession[pc] = p
	if !alreadyTracked {
		t.count.Inc()
	}

	// Update tracking for stream.
	byStreamMode, ok := t.presencesByStream[stream.Mode]
	if !ok {
		byStreamMode = make(map[PresenceStream]map[presenceCompact]*Presence)
		t.presencesByStream[stream.Mode] = byStreamMode
	}

	if byStream, ok := byStreamMode[stream]; !ok {
		byStream = make(map[presenceCompact]*Presence)
		byStream[pc] = p
		byStreamMode[stream] = byStream
	} else {
		byStream[pc] = p
	}

	t.Unlock()

	if !meta.Hidden || (alreadyTracked && !previousP.Meta.Hidden) {
		var joins []*Presence
		if !meta.Hidden {
			joins = []*Presence{p}
		}
		var leaves []*Presence
		if alreadyTracked && !previousP.Meta.Hidden {
			syncAtomic.StoreUint32(&previousP.Meta.Reason, uint32(runtime.PresenceReasonUpdate))
			leaves = []*Presence{previousP}
		}
		// Guaranteed joins and/or leaves are not empty or we wouldn't be inside this block.
		t.queueEvent(joins, leaves)
	}
	return true
}

func (t *LocalTracker) UntrackAllFromNode(node string, sessionID uuid.UUID, reason runtime.PresenceReason) {
	t.Lock()

	bySession, anyTracked := t.presencesBySession[sessionID]
	if !anyTracked {
		// Nothing tracked for the session.
		t.Unlock()
		return
	}

	leaves := make([]*Presence, 0, len(bySession))
	for pc, p := range bySession {
		// Update the tracking for stream.
		if byStreamMode := t.presencesByStream[pc.Stream.Mode]; len(byStreamMode) == 1 {
			// This is the only stream for this stream mode.
			if byStream := byStreamMode[pc.Stream]; len(byStream) == 1 {
				// This was the only presence in the only stream for this stream mode, discard the whole list.
				delete(t.presencesByStream, pc.Stream.Mode)
			} else {
				// There were other presences for the stream, drop just this one.
				delete(byStream, pc)
			}
		} else {
			// There are other streams for this stream mode.
			if byStream := byStreamMode[pc.Stream]; len(byStream) == 1 {
				// This was the only presence for the stream, discard the whole list.
				delete(byStreamMode, pc.Stream)
			} else {
				// There were other presences for the stream, drop just this one.
				delete(byStream, pc)
			}
		}

		// Check if there should be an event for this presence.
		if !p.Meta.Hidden {
			syncAtomic.StoreUint32(&p.Meta.Reason, uint32(reason))
			leaves = append(leaves, p)
		}

		t.count.Dec()
	}
	// Discard the tracking for session.
	delete(t.presencesBySession, sessionID)

	t.Unlock()
	if len(leaves) != 0 {
		t.queueEvent(nil, leaves)
	}
}

func (t *LocalTracker) UntrackByStreamFromNode(node string, stream PresenceStream) {
	// NOTE: Generates no presence notifications as everyone on the stream is going away all at once.
	t.Lock()

	byStream, anyTracked := t.presencesByStream[stream.Mode][stream]
	if !anyTracked {
		// Nothing tracked for the stream.
		t.Unlock()
		return
	}

	// Drop the presences from tracking for each session.
	for pc := range byStream {
		if bySession := t.presencesBySession[pc.ID.SessionID]; len(bySession) == 1 {
			// This is the only presence for that session, discard the whole list.
			delete(t.presencesBySession, pc.ID.SessionID)
		} else {
			// There were other presences for the session, drop just this one.
			delete(bySession, pc)
		}
		t.count.Dec()
	}

	// Discard the tracking for stream.
	if byStreamMode := t.presencesByStream[stream.Mode]; len(byStreamMode) == 1 {
		// This is the only stream for this stream mode.
		delete(t.presencesByStream, stream.Mode)
	} else {
		// There are other streams for this stream mode.
		delete(byStreamMode, stream)
	}

	t.Unlock()
}

func (t *LocalTracker) UntrackByModesFromNode(node string, sessionID uuid.UUID, modes map[uint8]struct{}, skipStream PresenceStream) {
	leaves := make([]*Presence, 0, 1)

	t.Lock()
	bySession, anyTracked := t.presencesBySession[sessionID]
	if !anyTracked {
		t.Unlock()
		return
	}

	for pc, p := range bySession {
		if _, found := modes[pc.Stream.Mode]; !found {
			// Not a stream mode we need to check.
			continue
		}
		if pc.Stream == skipStream {
			// Skip this stream based on input.
			continue
		}

		// Update the tracking for session.
		if len(bySession) == 1 {
			// This was the only presence for the session, discard the whole list.
			delete(t.presencesBySession, sessionID)
		} else {
			// There were other presences for the session, drop just this one.
			delete(bySession, pc)
		}
		t.count.Dec()

		// Update the tracking for stream.
		if byStreamMode := t.presencesByStream[pc.Stream.Mode]; len(byStreamMode) == 1 {
			// This is the only stream for this stream mode.
			if byStream := byStreamMode[pc.Stream]; len(byStream) == 1 {
				// This was the only presence in the only stream for this stream mode, discard the whole list.
				delete(t.presencesByStream, pc.Stream.Mode)
			} else {
				// There were other presences for the stream, drop just this one.
				delete(byStream, pc)
			}
		} else {
			// There are other streams for this stream mode.
			if byStream := byStreamMode[pc.Stream]; len(byStream) == 1 {
				// This was the only presence for the stream, discard the whole list.
				delete(byStreamMode, pc.Stream)
			} else {
				// There were other presences for the stream, drop just this one.
				delete(byStream, pc)
			}
		}

		if !p.Meta.Hidden {
			syncAtomic.StoreUint32(&p.Meta.Reason, uint32(runtime.PresenceReasonLeave))
			leaves = append(leaves, p)
		}
	}
	t.Unlock()

	if len(leaves) > 0 {
		t.queueEvent(nil, leaves)
	}
}

func (t *LocalTracker) UntrackByModes(sessionID uuid.UUID, modes map[uint8]struct{}, skipStream PresenceStream) {
	leaves := make([]*Presence, 0, 1)

	t.Lock()
	bySession, anyTracked := t.presencesBySession[sessionID]
	if !anyTracked {
		t.Unlock()
		return
	}

	for pc, p := range bySession {
		if _, found := modes[pc.Stream.Mode]; !found {
			// Not a stream mode we need to check.
			continue
		}
		if pc.Stream == skipStream {
			// Skip this stream based on input.
			continue
		}

		// Update the tracking for session.
		if len(bySession) == 1 {
			// This was the only presence for the session, discard the whole list.
			delete(t.presencesBySession, sessionID)
		} else {
			// There were other presences for the session, drop just this one.
			delete(bySession, pc)
		}
		t.count.Dec()

		// Update the tracking for stream.
		if byStreamMode := t.presencesByStream[pc.Stream.Mode]; len(byStreamMode) == 1 {
			// This is the only stream for this stream mode.
			if byStream := byStreamMode[pc.Stream]; len(byStream) == 1 {
				// This was the only presence in the only stream for this stream mode, discard the whole list.
				delete(t.presencesByStream, pc.Stream.Mode)
			} else {
				// There were other presences for the stream, drop just this one.
				delete(byStream, pc)
			}
		} else {
			// There are other streams for this stream mode.
			if byStream := byStreamMode[pc.Stream]; len(byStream) == 1 {
				// This was the only presence for the stream, discard the whole list.
				delete(byStreamMode, pc.Stream)
			} else {
				// There were other presences for the stream, drop just this one.
				delete(byStream, pc)
			}
		}

		if !p.Meta.Hidden {
			syncAtomic.StoreUint32(&p.Meta.Reason, uint32(runtime.PresenceReasonLeave))
			leaves = append(leaves, p)
		}
	}
	t.Unlock()

	if len(leaves) > 0 {
		t.queueEvent(nil, leaves)
	}

	if err := CC().NotifyUntrackByMode(sessionID, modes, skipStream); err != nil {
		t.logger.Error("Failed to notify UntrackByMode", zap.Error(err))
	}
}
