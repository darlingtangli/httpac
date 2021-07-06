package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

type ImpExt struct {
	Rewarded int `json:"rewarded"`
}

type Banner struct {
	Mimes []string `json:"mimes"`
	W     int      `json:"w"`
	H     int      `json:"h"`
	Ext   *ImpExt  `json:"ext"`
}

type Video struct {
	Mimes    []string `json:"mimes"`
	W        int      `json:"w"`
	H        int      `json:"h"`
	Rewarded int      `json:"rewarded"`
	Ext      *ImpExt  `json:"ext"`
}

type Audio struct {
	Mimes []string `json:"mimes"`
}

type Title struct {
	Len  int    `json:"int"`
	Text string `json:"text"`
}

type Image struct {
	Url   string   `json:"url"`
	W     int      `json:"w"`
	Wmin  int      `json:"wmin"`
	H     int      `json:"h"`
	Hmin  int      `json:"hmin"`
	Mimes []string `json:"mimes"`
}

type Data struct {
	Len   int    `json:"int"`
	Value string `json:"value"`
}

type Link struct {
	Url string `json:"url"`
}

type Native struct {
	Request string `json:"request"`
	Ver     string `json:"ver"`
}

type Asset struct {
	Id    int    `json:"id"`
	Title *Title `json:"title"`
	Image *Image `json:"image"`
	Video *Video `json:"video"`
	Data  *Data  `json:"data"`
	Link  *Link  `json:"link"`
}

type NativeRequest struct {
	Ver    string   `json:"ver"`
	Assets []*Asset `json:"assets"`
}

type NativeResponse struct {
	Ver    string   `json:"ver"`
	Assets []*Asset `json:"assets"`
}

type Imp struct {
	Instl       int     `json:"instl"`
	BidFloorCur string  `json:"bidfloorcur"`
	Banner      *Banner `json:"banner"`
	Video       *Video  `json:"video"`
	Audio       *Audio  `json:"audio"`
	Native      *Native `json:"native"`
	Ext         *ImpExt `json:"ext"`
}

type BidRequest struct {
	Imps []*Imp `json:"imp"`
}

type BidExt struct {
	Dsp string `json:"dsp"`
}

type Bid struct {
	Adm string  `json:"adm"`
	Ext *BidExt `json:"ext"`
}

type SeatBid struct {
	Bids []*Bid `json:"bid"`
}

type BidResponse struct {
	Code     int        `json:"code"`
	Seatbids []*SeatBid `json:"seatbid"`
}

var (
	files = map[string]*os.File{}
)

func file(name string) *os.File {
	_, ok := files[name]
	if !ok {
		files[name], _ = os.Create(name)
	}
	return files[name]
}

type NativeRequestStr struct {
	Native *NativeRequest `json:"native"`
}
type NativeResponseStr struct {
	Native *NativeResponse `json:"native"`
}

func dump2(t *trip) {
	buf := bufio.NewReader(bytes.NewBufferString(t.req))
	req, _ := http.ReadRequest(buf)
	defer req.Body.Close()
	if !strings.Contains(req.URL.Path, "GetAdOut") || req.URL.Query().Get("sspid") != "1105" {
		return
	}
	b, _ := ioutil.ReadAll(req.Body)
	bidReq := BidRequest{}
	json.Unmarshal(b, &bidReq)
	if len(bidReq.Imps) == 0 {
		fmt.Printf("bidReq.Imps empty: %s\n", b)
		return
	}

	var o io.Writer
	imp := bidReq.Imps[0]
	if imp.Video != nil {
		o = file("video.txt")
	} else if imp.Banner != nil {
		o = file("banner.txt")
	} else {
		o = file("native.txt")
	}

	//buf = bufio.NewReader(bytes.NewBufferString(t.rsp))
	//rsp, _ := http.ReadResponse(buf, nil)
	//defer rsp.Body.Close()
	//b, _ = ioutil.ReadAll(rsp.Body)
	//bidRsp := BidResponse{}
	//json.Unmarshal(b, &bidRsp)
	//if bidRsp.Code != 0 {
	//	return
	//}
	//if len(bidRsp.Seatbids) == 0 {
	//	fmt.Printf("bidRsp.Seatbids empty: %s\n", b)
	//	return
	//}
	//if len(bidRsp.Seatbids[0].Bids) == 0 {
	//	fmt.Printf("bidRsp.Seatbids[0].Bids empty: %s\n", b)
	//	return
	//}
	//bid := bidRsp.Seatbids[0].Bids[0]
	//if len(bid.Adm) == 0 {
	//	fmt.Printf("bidRsp.Seatbids[0].Bids empty: %s\n", b)
	//	return
	//}

	//var o io.Writer
	//if strings.HasPrefix(bid.Adm, "{") {
	//	var nativeRsp *NativeResponse
	//	if strings.Contains(bid.Adm, `{"native"`) {
	//		nativeRspTmp := &NativeResponseStr{}
	//		json.Unmarshal([]byte(bid.Adm), nativeRspTmp)
	//		nativeRsp = nativeRspTmp.Native
	//	} else {
	//		nativeRspTmp := &NativeResponse{}
	//		json.Unmarshal([]byte(bid.Adm), nativeRspTmp)
	//		nativeRsp = nativeRspTmp
	//	}
	//	name := "native_v" + nativeRsp.Ver
	//	video := false
	//	for _, a := range nativeRsp.Assets {
	//		if a.Video != nil {
	//			video = true
	//			break
	//		}
	//	}

	//	if video {
	//		o = file(name + "_video.txt")
	//	} else {
	//		o = file(name + ".txt")
	//	}
	//} else if strings.HasPrefix(bid.Adm, "<!DOCTYPE html") {
	//	o = file("banner.txt")
	//} else {
	//	o = file("video.txt")
	//}

	fmt.Fprintf(o, "\n\n######### local: %v remote: %v\n", t.ep.local, t.ep.remote)
	fmt.Fprintf(o, t.req)
	fmt.Fprintf(o, "\n### RSP\n%s", t.rsp)

}
