package core

import (
	"fmt"
	"sync"

	"github.com/ericxtang/m3u8"
	"github.com/golang/glog"
	"github.com/livepeer/go-livepeer/drivers"
	ffmpeg "github.com/livepeer/lpms/ffmpeg"
	"github.com/livepeer/lpms/stream"
)

const LIVE_LIST_LENGTH uint = 6

//	PlaylistManager manages playlists and data for one video stream, backed by one object storage.
type PlaylistManager interface {
	ManifestID() ManifestID
	// Implicitly creates master and media playlists
	// Inserts in media playlist given a link to a segment
	InsertHLSSegment(streamID StreamID, seqNo uint64, uri string, duration float64) error

	GetHLSMasterPlaylist() *m3u8.MasterPlaylist

	GetHLSMediaPlaylist(streamID StreamID) *m3u8.MediaPlaylist

	GetOSSession() drivers.OSSession

	Cleanup()
}

type BasicPlaylistManager struct {
	storageSession drivers.OSSession
	manifestID     ManifestID
	// Live playlist used for broadcasting
	masterPList *m3u8.MasterPlaylist
	mediaLists  map[StreamID]*m3u8.MediaPlaylist
	mapSync     *sync.RWMutex
}

// NewBasicPlaylistManager create new BasicPlaylistManager struct
func NewBasicPlaylistManager(manifestID ManifestID,
	storageSession drivers.OSSession) *BasicPlaylistManager {

	bplm := &BasicPlaylistManager{
		storageSession: storageSession,
		manifestID:     manifestID,
		masterPList:    m3u8.NewMasterPlaylist(),
		mediaLists:     make(map[StreamID]*m3u8.MediaPlaylist),
		mapSync:        &sync.RWMutex{},
	}
	return bplm
}

func (mgr *BasicPlaylistManager) ManifestID() ManifestID {
	return mgr.manifestID
}

func (mgr *BasicPlaylistManager) Cleanup() {
	mgr.storageSession.EndSession()
}

func (mgr *BasicPlaylistManager) GetOSSession() drivers.OSSession {
	return mgr.storageSession
}

func (mgr *BasicPlaylistManager) isRightStream(streamID StreamID) error {
	mid, err := streamID.ManifestIDFromStreamID()
	if err != nil {
		return err
	}
	if mid != mgr.manifestID {
		return fmt.Errorf("Wrong sream id (%s), should be %s", streamID, mid)
	}
	return nil
}

func (mgr *BasicPlaylistManager) hasSegment(mpl *m3u8.MediaPlaylist, seqNum uint64) bool {
	for _, seg := range mpl.Segments {
		if seg != nil && seg.SeqId == seqNum {
			return true
		}
	}
	return false
}

func (mgr *BasicPlaylistManager) getPL(streamID StreamID) *m3u8.MediaPlaylist {
	mgr.mapSync.RLock()
	mpl := mgr.mediaLists[streamID]
	mgr.mapSync.RUnlock()
	return mpl
}

func (mgr *BasicPlaylistManager) createPL(streamID StreamID) *m3u8.MediaPlaylist {
	mpl, err := m3u8.NewMediaPlaylist(LIVE_LIST_LENGTH, LIVE_LIST_LENGTH)
	if err != nil {
		glog.Error(err)
		return nil
	}
	mgr.mapSync.Lock()
	mgr.mediaLists[streamID] = mpl
	mgr.mapSync.Unlock()
	sid := string(streamID)
	vParams := ffmpeg.VideoProfileToVariantParams(ffmpeg.VideoProfileLookup[streamID.GetRendition()])
	url := sid + ".m3u8"
	mgr.masterPList.Append(url, mpl, vParams)
	return mpl
}

func (mgr *BasicPlaylistManager) InsertHLSSegment(streamID StreamID, seqNo uint64, uri string,
	duration float64) error {

	if err := mgr.isRightStream(streamID); err != nil {
		return err
	}
	mpl := mgr.getPL(streamID)
	if mpl == nil {
		mpl = mgr.createPL(streamID)
	}
	return mgr.addToMediaPlaylist(uri, seqNo, duration, mpl)
}

func (mgr *BasicPlaylistManager) mediaSegmentFromURI(uri string, seqNo uint64, duration float64) *m3u8.MediaSegment {
	mseg := new(m3u8.MediaSegment)
	mseg.URI = uri
	mseg.SeqId = seqNo
	mseg.Duration = duration
	return mseg
}

func (mgr *BasicPlaylistManager) addToMediaPlaylist(uri string, seqNo uint64, duration float64,
	mpl *m3u8.MediaPlaylist) error {

	mseg := mgr.mediaSegmentFromURI(uri, seqNo, duration)
	if mpl.Count() >= mpl.WinSize() {
		mpl.Remove()
	}
	if mpl.Count() == 0 {
		mpl.SeqNo = mseg.SeqId
	}
	return mpl.AppendSegment(mseg)
}

func (mgr *BasicPlaylistManager) makeMediaSegment(seg *stream.HLSSegment, url string) *m3u8.MediaSegment {
	mseg := new(m3u8.MediaSegment)
	mseg.URI = url
	mseg.Duration = seg.Duration
	mseg.Title = seg.Name
	mseg.SeqId = seg.SeqNo
	return mseg
}

// GetHLSMasterPlaylist ..
func (mgr *BasicPlaylistManager) GetHLSMasterPlaylist() *m3u8.MasterPlaylist {
	return mgr.masterPList
}

// GetHLSMediaPlaylist ...
func (mgr *BasicPlaylistManager) GetHLSMediaPlaylist(streamID StreamID) *m3u8.MediaPlaylist {
	return mgr.mediaLists[streamID]
}
