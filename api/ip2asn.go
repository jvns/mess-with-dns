package main

import (
    "bufio"
    "bytes"
    "errors"
    "net"
    "os"
    "strconv"
    "strings"
)

type Ranges struct {
    IPv4Ranges []IPRange
    IPv6Ranges []IPRange
}

type IPRange struct {
    StartIP net.IP
    EndIP net.IP
    Num int
    Name string
    Country string
}

func parseInt(s string) int {
    i, err := strconv.Atoi(s)
    if err != nil {
        return 0
    }
    return i
}

func ReadRanges() (Ranges, error) {
    ipv4Ranges, err := ReadASNs("ip2asn-v4.txt")
    if err != nil {
        return Ranges{}, err
    }
    ipv6Ranges, err := ReadASNs("ip2asn-v6.txt")
    if err != nil {
        return Ranges{}, err
    }
    return Ranges{
        IPv4Ranges: ipv4Ranges,
        IPv6Ranges: ipv6Ranges,
    }, nil
}

func (ranges Ranges) FindASN(ip net.IP) (IPRange, error) {
    if ip.To4() != nil {
        return FindASN(ranges.IPv4Ranges, ip)
    } else {
        return FindASN(ranges.IPv6Ranges, ip)
    }
}

func FindASN(lines []IPRange, ip net.IP) (IPRange, error) {
    // binary search
    start := 0
    end := len(lines) - 1
    for start <= end {
        mid := (start + end) / 2
        // check if it's between StartIP and EndIP
        if bytes.Compare(ip, lines[mid].StartIP) >= 0 && bytes.Compare(ip, lines[mid].EndIP) <= 0 {
            return lines[mid], nil
        } else if bytes.Compare(ip, lines[mid].StartIP) < 0 {
            end = mid - 1
        } else {
            start = mid + 1
        }
    }
    return IPRange{}, errors.New("not found")
}

func ReadASNs(filename string) ([]IPRange, error) {
    f, err := os.Open(filename)
    if err != nil {
        return nil, err
    }
    defer f.Close()
    // read lines
    scanner := bufio.NewScanner(f)
    var lines []IPRange
    for scanner.Scan() {
        line := scanner.Text()
        // split line
        fields := strings.Split(line, "\t")
        // parse fields
        IPRange := IPRange{
            StartIP: net.ParseIP(fields[0]),
            EndIP: net.ParseIP(fields[1]),
            Num: parseInt(fields[2]),
            Country: fields[3],
            Name: fields[4],
        }
        lines = append(lines, IPRange)
    }
    return lines, nil
}
