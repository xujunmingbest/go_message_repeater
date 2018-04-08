package main

import (
	"strings"
	"encoding/xml"
	"io/ioutil"
	"log"
)

func ConfigToMap(){
	repeater_map_lock.Lock()
	defer repeater_map_lock.Unlock()
	repeater_register_map_lock.Lock()
	defer repeater_register_map_lock.Unlock()

	for _,v := range CR.Mrs{
		repeater_register_map[v.Pair_mac1] = ""
		repeater_register_map[v.Pair_mac2] = ""
		r := strings.Compare(v.Pair_mac1,v.Pair_mac2)
		if  r < 0{
			repeater_map[v.Pair_mac1+v.Pair_mac2] = ""
		}else {
			repeater_map[v.Pair_mac2+v.Pair_mac1] = ""
		}
	}
}


func ConfigReload(){




}

type ConfigResult struct {
	XMLName xml.Name `xml:"message_repeater"`
	Repeater_max_number int `xml:"repeater_max_number"`
	Mrs []Repeater_pair `xml:"repeater_pair"`
}

type Repeater_pair struct {
	Pair_mac1 string `xml:"pair1_mac1"` //默认是 mrcMacPair_mac1
	Pair_mac2 string `xml:"pair1_mac2"` //Pair_mac2
}


func ConfigLoad() {
	content, err := ioutil.ReadFile("E:\\go\\message_repeater\\config.xml")
	if err != nil {
		log.Fatal(err)
	}
	err = xml.Unmarshal(content, &CR)
	if err != nil {
		log.Fatal(err)
	}
}

