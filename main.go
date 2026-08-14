package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"os"
	"unsafe"

	"golang.org/x/sys/unix"
)

const (
	IFF_TUN   = 0x0001
	IFF_NO_PI = 0x1000
	TUNSETIFF = 0x400454ca
)

type ifreq struct {
	Name  [16]byte
	Flags uint16
	pad   [22]byte
}

func createTUN(name string) (*os.File, error) {
	fd, err := unix.Open("/dev/net/tun", unix.O_RDWR, 0)
	if err != nil {
		return nil, err
	}
	var req ifreq
	copy(req.Name[:], name)
	req.Flags = IFF_TUN | IFF_NO_PI
	_, _, errno := unix.Syscall(unix.SYS_IOCTL, uintptr(fd), uintptr(TUNSETIFF), uintptr(unsafe.Pointer(&req)))
	if errno != 0 {
		return nil, errno
	}
	return os.NewFile(uintptr(fd), name), nil
}

func parseIPv4(packet []byte) {
	// extract version (top nibble of byte 0)
	version := packet[0] >> 4

	// if version != 4, print a skip message and return
	if version != 4 {
		fmt.Println("IPV6: returning")
		return
	}
	// extract IHL (bottom nibble of byte 0, times 4)
	ihl := (packet[0] & 0x0F) * 4
	// extract total length (bytes 2-3, BigEndian uint16)
	totalLength := binary.BigEndian.Uint16(packet[2:4])
	// extract protocol (byte 9)
	protocol := packet[9]
	// extract source IP (bytes 12-16) and dest IP (bytes 16-20)
	srcIP := net.IP(packet[12:16])
	dstIP := net.IP(packet[16:20])

	fmt.Printf("IPv%d ihl=%d totalLen=%d proto=%d src=%s dst=%s\n", version, ihl, totalLength, protocol, srcIP, dstIP)
}

func checksum(data []byte) uint16 {
	var sum uint32
	for i := 0; i < len(data)-1; i += 2 {
		sum += uint32(data[i])<<8 | uint32(data[i+1])
	}
	if len(data)%2 == 1 {
		sum += uint32(data[len(data)-1]) << 8
	}
	for sum>>16 != 0 {
		sum = (sum & 0xFFFF) + (sum >> 16)
	}
	return ^uint16(sum)
}

func handleICMP(packet []byte, tun *os.File) {
	ihl := (packet[0] & 0x0F) * 4
	protocol := packet[9]

	if protocol != 1 {
		return
	}

	if len(packet) < int(ihl)+8 {
		return // too short to be a valid ICMP echo, skip
	}

	icmp := packet[ihl:]
	if icmp[0] != 8 {
		return
	}
	for i := 0; i < 4; i++ {
		packet[12+i], packet[16+i] = packet[16+i], packet[12+i]
	}
	icmp[0] = 0
	icmp[2] = 0
	icmp[3] = 0

	newChecksum := checksum(icmp)
	binary.BigEndian.PutUint16(icmp[2:4], newChecksum)

	_, err := tun.Write(packet)
	if err != nil {
		fmt.Println("write error:", err)
	}
}

func main() {
	tun, err := createTUN("tun0")
	if err != nil {
		log.Fatal("failed to create TUN device:", err)
	}
	defer tun.Close()

	fmt.Println("TUN device tun0 created. In another terminal, run:")
	fmt.Println("  sudo ip addr add 10.0.0.1/24 dev tun0")
	fmt.Println("  sudo ip link set tun0 up")
	fmt.Println("Then: ping 10.0.0.2")

	buf := make([]byte, 1500)
	for {
		n, err := tun.Read(buf)
		if err != nil {
			log.Fatal("read error:", err)
		}
		parseIPv4(buf[:n])
		handleICMP(buf[:n], tun)
	}
}
