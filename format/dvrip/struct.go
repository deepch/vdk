package dvrip

import (
	"crypto/md5"
	"encoding/binary"
	"time"
)

var magicEnd = [2]byte{0x0A, 0x00}

const alnum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

type Payload struct {
	Head           byte
	Version        byte
	_              byte
	_              byte
	Session        int32
	SequenceNumber int32
	_              byte
	_              byte
	MsgID          int16
	BodyLength     int32
}
type LoginResp struct {
	AliveInterval int    `json:"AliveInterval"`
	ChannelNum    int    `json:"ChannelNum"`
	DeviceType    string `json:"DeviceType "`
	ExtraChannel  int    `json:"ExtraChannel"`
	Ret           int    `json:"Ret"`
	SessionID     string `json:"SessionID"`
}

type requestCode uint16

const (
	codeLogin            requestCode = 1000
	codeKeepAlive        requestCode = 1006
	codeSystemInfo       requestCode = 1020
	codeNetWorkNetCommon requestCode = 1042
	codeGeneral          requestCode = 1042
	codeChannelTitle     requestCode = 1046
	codeSystemFunction   requestCode = 1360
	codeEncodeCapability requestCode = 1360
	codeOPPTZControl     requestCode = 1400
	codeOPMonitor        requestCode = 1413
	codeOPTalk           requestCode = 1434
	codeOPTimeSetting    requestCode = 1450
	codeOPMachine        requestCode = 1450
	codeOPTimeQuery      requestCode = 1452
	codeAuthorityList    requestCode = 1470
	codeUsers            requestCode = 1472
	codeGroups           requestCode = 1474
	codeAddGroup         requestCode = 1476
	codeModifyGroup      requestCode = 1478
	codeDelGroup         requestCode = 1480
	codeAddUser          requestCode = 1482
	codeModifyUser       requestCode = 1484
	codeDelUser          requestCode = 1486
	codeModifyPassword   requestCode = 1488
	codeAlarmSet         requestCode = 1500
	codeOPNetAlarm       requestCode = 1506
	codeAlarmInfo        requestCode = 1504
	codeOPSendFile       requestCode = 1522
	codeOPSystemUpgrade  requestCode = 1525
	codeOPNetKeyboard    requestCode = 1550
	codeOPSNAP           requestCode = 1560
	codeOPMailTest       requestCode = 1636
)

type statusCode int

const (
	statusOK                                  statusCode = 100
	statusUnknownError                        statusCode = 101
	statusUnsupportedVersion                  statusCode = 102
	statusRequestNotPermitted                 statusCode = 103
	statusUserAlreadyLoggedIn                 statusCode = 104
	statusUserIsNotLoggedIn                   statusCode = 105
	statusUsernameOrPasswordIsIncorrect       statusCode = 106
	statusUserDoesNotHaveNecessaryPermissions statusCode = 107
	statusPasswordIsIncorrect                 statusCode = 203
	statusStartOfUpgrade                      statusCode = 511
	statusUpgradeWasNotStarted                statusCode = 512
	statusUpgradeDataErrors                   statusCode = 513
	statusUpgradeError                        statusCode = 514
	statusUpgradeSuccessful                   statusCode = 515
)

var statusCodes = map[statusCode]string{
	statusOK:                                  "OK",
	statusUnknownError:                        "Unknown error",
	statusUnsupportedVersion:                  "Unsupported version",
	statusRequestNotPermitted:                 "Request not permitted",
	statusUserAlreadyLoggedIn:                 "User already logged in",
	statusUserIsNotLoggedIn:                   "User is not logged in",
	statusUsernameOrPasswordIsIncorrect:       "Username or password is incorrect",
	statusUserDoesNotHaveNecessaryPermissions: "User does not have necessary permissions",
	statusPasswordIsIncorrect:                 "Password is incorrect",
	statusStartOfUpgrade:                      "Start of upgrade",
	statusUpgradeWasNotStarted:                "Upgrade was not started",
	statusUpgradeDataErrors:                   "Upgrade data errors",
	statusUpgradeError:                        "Upgrade error",
	statusUpgradeSuccessful:                   "Upgrade successful",
}

var requestCodes = map[requestCode]string{
	codeOPMonitor:     "OPMonitor",
	codeOPTimeSetting: "OPTimeSetting",
}

type MetaInfo struct {
	Width    int
	Height   int
	Datetime time.Time
	FPS      int
	Frame    string
	Type     string
}

type Frame struct {
	Data []byte
	Meta MetaInfo
}

//sofiaHash func
func sofiaHash(password string) string {
	digest := md5.Sum([]byte(password))
	hash := make([]byte, 0, 8)
	for i := 1; i < len(digest); i += 2 {
		sum := int(digest[i-1]) + int(digest[i])
		hash = append(hash, alnum[sum%len(alnum)])
	}
	return string(hash)
}

//parseMediaType func
func parseMediaType(dataType uint32, mediaCode byte) string {
	switch dataType {
	case 0x1FC, 0x1FD:
		switch mediaCode {
		case 1:
			return "MPEG4"
		case 2:
			return "H264"
		case 3:
			return "H265"
		}
	case 0x1F9:
		if mediaCode == 1 || mediaCode == 6 {
			return "info"
		}
	case 0x1FA:
		if mediaCode == 0xE {
			return "PCM_ALAW"
		}
	case 0x1FE:
		if mediaCode == 0 {
			return "JPEG"
		}
	default:
		return "unknown"
	}

	return "unexpected"
}

//parseDatetime func
func parseDatetime(value uint32) time.Time {
	second := int(value & 0x3F)
	minute := int((value & 0xFC0) >> 6)
	hour := int((value & 0x1F000) >> 12)
	day := int((value & 0x3E0000) >> 17)
	month := int((value & 0x3C00000) >> 22)
	year := int(((value & 0xFC000000) >> 26) + 2000)

	return time.Date(year, time.Month(month), day, hour, minute, second, 0, time.UTC)
}

//binSize func
func binSize(val int) []byte {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(val))
	return buf
}
