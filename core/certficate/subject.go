/*
   This file is part of go-palletone.
   go-palletone is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.
   go-palletone is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.
   You should have received a copy of the GNU General Public License
   along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
*/
package certficate

import (
	"crypto/x509/pkix"
	"flag"
	"fmt"
	"net"
	"strings"

	errgo "gopkg.in/errgo.v1"
)

var (
	subject         nameVar
	subjectAltDNS   namesVar
	subjectAltEmail namesVar
	subjectAltIP    ipsVar
)

func init() {
	flag.Var(&subject, "subject", "`name` to which the certificate is issued.")
	flag.Var(&subjectAltDNS, "subject-alt-name", "alternative `name` of the subject.")
	flag.Var(&subjectAltEmail, "subject-alt-email", "alternative `email address` of the subject.")
	flag.Var(&subjectAltIP, "subject-alt-ip", "alternative `IP address` of the subject.")
}

func Subject() pkix.Name {
	return subject.name
}

func DNSNames() []string {
	return []string(subjectAltDNS)
}

func EmailAddresses() []string {
	return []string(subjectAltEmail)
}

func IPAddresses() []net.IP {
	return []net.IP(subjectAltIP)
}

type namesVar []string

func (v *namesVar) Set(s string) error {
	for _, s := range strings.Split(s, ",") {
		*v = append(*v, s)
	}
	return nil
}

func (v namesVar) String() string {
	return strings.Join([]string(v), ",")
}

type ipsVar []net.IP

func (v *ipsVar) Set(s string) error {
	for _, s := range strings.Split(s, ",") {
		ip := net.ParseIP(s)
		if ip == nil {
			return errgo.Newf("invalid IP address %q", s)
		}
		*v = append(*v, ip)
	}
	return nil
}

func (v ipsVar) String() string {
	ss := make([]string, len(v))
	for i, ip := range v {
		ss[i] = ip.String()
	}
	return strings.Join(ss, ",")
}

type nameVar struct {
	name pkix.Name
}

var nameParts = map[string]func(*pkix.Name, string) error{
	"C": func(n *pkix.Name, s string) error {
		n.Country = append(n.Country, s)
		return nil
	},
	"O": func(n *pkix.Name, s string) error {
		n.Organization = append(n.Organization, s)
		return nil
	},
	"OU": func(n *pkix.Name, s string) error {
		n.OrganizationalUnit = append(n.OrganizationalUnit, s)
		return nil
	},
	"L": func(n *pkix.Name, s string) error {
		n.Locality = append(n.Locality, s)
		return nil
	},
	"ST": func(n *pkix.Name, s string) error {
		n.Province = append(n.Province, s)
		return nil
	},
	"STREET": func(n *pkix.Name, s string) error {
		n.StreetAddress = append(n.StreetAddress, s)
		return nil
	},
	"PC": func(n *pkix.Name, s string) error {
		n.PostalCode = append(n.PostalCode, s)
		return nil
	},
	"CN": func(n *pkix.Name, s string) error {
		if n.CommonName != "" {
			return errgo.Newf("too many CN components")
		}
		n.CommonName = s
		return nil
	},
	"SERIALNUMBER": func(n *pkix.Name, s string) error {
		if n.SerialNumber != "" {
			return errgo.Newf("too many SERIALNUMBER components")
		}
		n.SerialNumber = s
		return nil
	},
}

func (v *nameVar) Set(s string) error {
	for _, t := range strings.Split(s, ",") {
		t := strings.TrimSpace(t)
		i := strings.Index(t, "=")
		if i == -1 {
			return errgo.Newf("invalid name component %q", t)
		}
		if f, ok := nameParts[t[:i]]; ok {
			f(&v.name, t[i+1:])
			continue
		}
		return errgo.Newf("unrecognized name component %q", t[:i])
	}
	return nil
}

func (v nameVar) String() string {
	parts := make([]string, 0, 9)
	for _, c := range v.name.Country {
		parts = append(parts, fmt.Sprintf("C=%s", c))
	}
	for _, p := range v.name.Province {
		parts = append(parts, fmt.Sprintf("ST=%s", p))
	}
	for _, l := range v.name.Locality {
		parts = append(parts, fmt.Sprintf("L=%s", l))
	}
	for _, s := range v.name.StreetAddress {
		parts = append(parts, fmt.Sprintf("STREET=%s", s))
	}
	for _, p := range v.name.PostalCode {
		parts = append(parts, fmt.Sprintf("PC=%s", p))
	}
	for _, o := range v.name.Organization {
		parts = append(parts, fmt.Sprintf("O=%s", o))
	}
	for _, o := range v.name.OrganizationalUnit {
		parts = append(parts, fmt.Sprintf("OU=%s", o))
	}
	if v.name.CommonName != "" {
		parts = append(parts, fmt.Sprintf("CN=%s", v.name.CommonName))
	}
	if v.name.SerialNumber != "" {
		parts = append(parts, fmt.Sprintf("SERIALNUMBER=%s", v.name.SerialNumber))
	}
	return strings.Join(parts, ", ")
}
