package main

import "fmt"

type Request struct {
	code       uint
	valueInt   int
	valueFloat float64
	origin     uint
}

var requestChannel chan Request
var requestChannelWait chan bool
var requestChannelGo chan bool

func synchronizer() {
	for {
		// open request channels (must be here for the first request, when the program starts)
		requestChannelGo <- true
		// wait for a reqeust
		request := <-requestChannel
		// tell other requesters to wait
		requestChannelWait <- true

		fmt.Println("INFO[synchronizer]: new request\n- code:", codeToString(request.code), "\n- int value:", request.valueInt,
			"\n- float value:", request.valueFloat, "\n- origin:", originToString(request.origin))

		switch request.code {
		case SEEK:
			seek(request.valueInt)
		case VOL:
			/*
				changing the volume via anything will trigger the OnChanged function in slider widget.
				in case the request came from a button or shortcut, we don't want to trigger the change
				from OnChanged function. after completing the current request, we will briefly allow the new requests to come,
				and then shut them down immediately. because changing the volume with the current request triggered OnChanged,
				a new request was sent as soon as we changed the volume. therefore, the first next request is the volume request
				coming from the slider - the exact one we want to skip.
			*/
			if request.origin == BTN || request.origin == SC {
				playerCtrl.changeVolume(request.valueFloat)
				fmt.Println("trying to catch the slider request...")
				requestChannelGo <- true
				requestChannelWait <- true
				select {
				case newRequest := <-requestChannel:
					fmt.Println("INFO[synchronizer]: followup request\n- code:", codeToString(request.code), "\n- int value:", request.valueInt,
						"\n- float value:", request.valueFloat, "\n- origin:", originToString(request.origin))
					// if the new request isn't a volume request, or is a volume request, but not from the slider, we will send it again
					// because it's not a request that we should skip.
					if newRequest.code != VOL || newRequest.origin != SLIDER {
						sendRequest(newRequest)
					}
					// otherwise, just continue (thus skipping the volume + slider request.)
				default:
				}
				// we always skip the unwanted volume + slider request, as explained previously. therefore, if we catch a volume + slider
				// request here, that means the slider was actually used to change the volume.
			} else if request.origin == SLIDER {
				playerCtrl.changeVolume(request.valueFloat)
			}
		default:
		}

	}
}

func sendRequest(request Request) {
	select {
	case <-requestChannelWait:
		<-requestChannelGo
		requestChannel <- request
	default:
		requestChannel <- request
	}
}

const (
	// request code
	SEEK uint = 0
	VOL  uint = 1
	// request origin
	SC     uint = 10
	BTN    uint = 11
	SLIDER uint = 12
)

func codeToString(code uint) string {
	switch code {
	case SEEK:
		return "SEEK"
	case VOL:
		return "VOL"
	default:
	}
	return "NO CODE"
}

func originToString(origin uint) string {
	switch origin {
	case SC:
		return "SC"
	case BTN:
		return "BTN"
	case SLIDER:
		return "SLIDER"
	default:
	}

	return "NO ORIGIN"
}
