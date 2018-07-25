package iplookup

import (
	"net"
	"net/http"
	"strings"
)

type NetId struct {
	InternalIPString string
	ExternalIPString string
	MacAddrString    string
}

func GetOutboundIPString() (string, error) {
	// SO:23558425
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String(), nil
}

func GetExternalIPString() (string, error) {
	resp, err := http.Get("http://ipv4.myexternalip.com/raw")
	if err != nil {
		//errLog.Println("failed to get external IP")
		return "", err
	}
	data := make([]byte, 1024)
	n, err := resp.Body.Read(data)
	//if err != nil {
	//errLog.Println("failed to get external IP")
	//return "", err
	//}
	//errLog.Println("external IP: ", string(data[0:n]))
	defer resp.Body.Close()
	ipstring := string(data[0:n])
	return strings.TrimRight(ipstring, "\n"), nil

}

func GetFirstMACAddrString() (string, error) {
	mac := "not_available"
	ifaces, err := net.Interfaces()
	if err != nil {
		return mac, err
	}
	for _, iface := range ifaces {
		mac := iface.HardwareAddr.String()
		if mac != "" {
			return mac, nil
		}
	}
	return mac, nil
}

func FindNetId() NetId {
	internalIPString, _ := GetOutboundIPString()
	externalIPString, _ := GetExternalIPString()
	macAddrString, _ := GetFirstMACAddrString()
	return NetId{
		InternalIPString: internalIPString,
		ExternalIPString: externalIPString,
		MacAddrString:    macAddrString,
	}
}
