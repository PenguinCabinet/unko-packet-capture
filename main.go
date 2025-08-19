package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/urfave/cli/v3"
)

func containsIP(ip_list []net.IP, ip net.IP) bool {
	for _, l := range ip_list {
		if l.Equal(ip) {
			return true
		}
	}
	return false
}

type packet_send_or_received_t int

const (
	packet_send_dist packet_send_or_received_t = iota
	packet_received_dist
)

func send_or_received(packet gopacket.Packet, hardware_address net.HardwareAddr, localIP_list []net.IP) packet_send_or_received_t {
	ethernet_layer := packet.Layer(layers.LayerTypeEthernet)
	ip_layer := packet.Layer(layers.LayerTypeIPv4)
	if ethernet_layer != nil && ip_layer != nil {
		eth := ethernet_layer.(*layers.Ethernet)
		ip := ip_layer.(*layers.IPv4)

		if eth.SrcMAC.String() == hardware_address.String() || ip.SrcIP.Equal(localIP_list[0]) {
			return packet_send_dist
		} else if eth.DstMAC.String() == hardware_address.String() || containsIP(localIP_list, ip.DstIP) {
			return packet_received_dist
		}
	}
	return packet_received_dist
}

type unko_t struct {
	pos  int
	dist int
}

func print_unko(unko_list [][]unko_t, unko string) {
	fmt.Print("\033[H\033[2J")
	for y, e := range unko_list {
		for i := 0; i < len(e); i++ {
			fmt.Printf("\x1b[%d;%dH", y, e[i].pos)
			fmt.Print(unko)
		}
	}
}

func move_unko(unko_list [][]unko_t, size int) {
	for y, _ := range unko_list {
		temp := []unko_t{}
		for x, _ := range unko_list[y] {
			unko_list[y][x].pos += unko_list[y][x].dist
			if 0 < unko_list[y][x].pos && unko_list[y][x].pos <= size {
				temp = append(temp, unko_list[y][x])
			}
		}

		unko_list[y] = temp
	}
}

func main() {
	cmd := &cli.Command{
		Commands: []*cli.Command{
			{
				Name:    "pcap devs",
				Aliases: []string{"pd"},
				Usage:   "Get pcap devices list.",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					devs, err := pcap.FindAllDevs()
					if err != nil {
						log.Fatalln(err)
					}

					for _, dev := range devs {
						fmt.Printf("Name: %s\n", dev.Name)
						fmt.Printf("Description: %s\n", dev.Description)
						fmt.Println("")
					}
					return nil
				},
			},
			{
				Name:    "net devs",
				Aliases: []string{"nd"},
				Usage:   "Get net devices list.",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					interface_list, err := net.Interfaces()
					if err != nil {
						log.Fatalln(err)
					}

					for _, e := range interface_list {
						if e.HardwareAddr != nil && len(e.HardwareAddr) > 0 {
							fmt.Printf("Name: %s\n", e.Name)
						}
					}
					return nil
				},
			},
		},
		Usage: "unko packet capture",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "pdev",
				Value: "",
				Usage: "pcap devices name",
			},

			&cli.StringFlag{
				Name:  "ndev",
				Value: "",
				Usage: "net devices name",
			},
			&cli.IntFlag{
				Name:  "size",
				Value: 70,
				Usage: "screen size",
			},
			&cli.IntFlag{
				Name:  "frame",
				Value: 300,
				Usage: "frame time(ms)",
			},
			&cli.StringFlag{
				Name:  "unko",
				Value: "💩",
				Usage: "Are you really going to change it?",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {

			pdev := cmd.String("pdev")
			ndev := cmd.String("ndev")
			size := cmd.Int("size")
			unko := cmd.String("unko")

			dev_interface, err := net.InterfaceByName(ndev)
			if err != nil {
				log.Fatalln(dev_interface)
			}

			addr_list, err := dev_interface.Addrs()
			if err != nil {
				log.Fatalln(dev_interface)
			}

			var localIP_list []net.IP
			for _, addr := range addr_list {
				if ipNet, ok := addr.(*net.IPNet); ok {
					localIP_list = append(localIP_list, ipNet.IP)
				}
			}

			var hardware_addr net.HardwareAddr = nil
			hardware_addr = dev_interface.HardwareAddr

			Handle, err := pcap.OpenLive(pdev, 1024, false, 10*time.Second)
			if err != nil {
				log.Fatal(err)
			}
			defer Handle.Close()

			unko_list := make([][]unko_t, size)

			packet_source := gopacket.NewPacketSource(Handle, Handle.LinkType())

			go func() {
				for {
					print_unko(unko_list, unko)
					move_unko(unko_list, size)
					time.Sleep(time.Duration(cmd.Int("frame")) * time.Millisecond)
				}
			}()

			for packet := range packet_source.Packets() {
				packet_dist := send_or_received(packet, hardware_addr, localIP_list)

				rand_int := rand.Intn(len(unko_list))

				if packet_dist == packet_send_dist {
					unko_list[rand_int] = append(unko_list[rand_int], unko_t{pos: 0, dist: 1})
				} else {
					unko_list[rand_int] = append(unko_list[rand_int], unko_t{pos: size, dist: -1})
				}

				time.Sleep(100 * time.Millisecond)
			}

			return nil
		},
	}
	(cmd).Run(context.Background(), os.Args)
}
