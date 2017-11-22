// +build !windows

package termcodes

import (
	"os"
	"syscall"

	"golang.org/x/crypto/ssh"
)

// struct termios {
// 	tcflag_t c_iflag;   /* input modes */
// 	tcflag_t c_oflag;   /* output modes */
// 	tcflag_t c_cflag;   /* control modes */
// 	tcflag_t c_lflag;   /* local modes */
// 	cc_t c_cc[NCCS];    /* special characters */
// 	speed_t c_ispeed;
// 	speed_t c_ospeed;
// };

type Setter interface {
	Set(pty *os.File, termios *syscall.Termios, value uint32) error
}

var TermAttrSetters map[uint8]Setter = map[uint8]Setter{
	ssh.VINTR:    &ccSetter{Character: syscall.VINTR},
	ssh.VQUIT:    &ccSetter{Character: syscall.VQUIT},
	ssh.VERASE:   &ccSetter{Character: syscall.VERASE},
	ssh.VKILL:    &ccSetter{Character: syscall.VKILL},
	ssh.VEOF:     &ccSetter{Character: syscall.VEOF},
	ssh.VEOL:     &ccSetter{Character: syscall.VEOL},
	ssh.VEOL2:    &ccSetter{Character: syscall.VEOL2},
	ssh.VSTART:   &ccSetter{Character: syscall.VSTART},
	ssh.VSTOP:    &ccSetter{Character: syscall.VSTOP},
	ssh.VSUSP:    &ccSetter{Character: syscall.VSUSP},
	ssh.VDSUSP:   &nopSetter{},
	ssh.VREPRINT: &ccSetter{Character: syscall.VREPRINT},
	ssh.VWERASE:  &ccSetter{Character: syscall.VWERASE},
	ssh.VLNEXT:   &ccSetter{Character: syscall.VLNEXT},
	ssh.VFLUSH:   &nopSetter{},
	ssh.VSWTCH:   &nopSetter{},
	ssh.VSTATUS:  &nopSetter{},
	ssh.VDISCARD: &ccSetter{Character: syscall.VDISCARD},

	// Input modes
	ssh.IGNPAR:  &iflagSetter{Flag: syscall.IGNPAR},
	ssh.PARMRK:  &iflagSetter{Flag: syscall.PARMRK},
	ssh.INPCK:   &iflagSetter{Flag: syscall.INPCK},
	ssh.ISTRIP:  &iflagSetter{Flag: syscall.ISTRIP},
	ssh.INLCR:   &iflagSetter{Flag: syscall.INLCR},
	ssh.IGNCR:   &iflagSetter{Flag: syscall.IGNCR},
	ssh.ICRNL:   &iflagSetter{Flag: syscall.ICRNL},
	ssh.IUCLC:   &nopSetter{},
	ssh.IXON:    &iflagSetter{Flag: syscall.IXON},
	ssh.IXANY:   &iflagSetter{Flag: syscall.IXANY},
	ssh.IXOFF:   &iflagSetter{Flag: syscall.IXOFF},
	ssh.IMAXBEL: &iflagSetter{Flag: syscall.IMAXBEL},

	// Local modes
	ssh.ISIG:    &lflagSetter{Flag: syscall.ISIG},
	ssh.ICANON:  &lflagSetter{Flag: syscall.ICANON},
	ssh.XCASE:   &nopSetter{},
	ssh.ECHO:    &lflagSetter{Flag: syscall.ECHO},
	ssh.ECHOE:   &lflagSetter{Flag: syscall.ECHOE},
	ssh.ECHOK:   &lflagSetter{Flag: syscall.ECHOK},
	ssh.ECHONL:  &lflagSetter{Flag: syscall.ECHONL},
	ssh.NOFLSH:  &lflagSetter{Flag: syscall.NOFLSH},
	ssh.TOSTOP:  &lflagSetter{Flag: syscall.TOSTOP},
	ssh.IEXTEN:  &lflagSetter{Flag: syscall.IEXTEN},
	ssh.ECHOCTL: &lflagSetter{Flag: syscall.ECHOCTL},
	ssh.ECHOKE:  &lflagSetter{Flag: syscall.ECHOKE},
	ssh.PENDIN:  &lflagSetter{Flag: syscall.PENDIN},

	// Output modes
	ssh.OPOST:  &oflagSetter{Flag: syscall.OPOST},
	ssh.OLCUC:  &nopSetter{},
	ssh.ONLCR:  &oflagSetter{Flag: syscall.ONLCR},
	ssh.OCRNL:  &oflagSetter{Flag: syscall.OCRNL},
	ssh.ONOCR:  &oflagSetter{Flag: syscall.ONOCR},
	ssh.ONLRET: &oflagSetter{Flag: syscall.ONLRET},

	// Control modes
	ssh.CS7:    &cflagSetter{Flag: syscall.CS7},
	ssh.CS8:    &cflagSetter{Flag: syscall.CS8},
	ssh.PARENB: &cflagSetter{Flag: syscall.PARENB},
	ssh.PARODD: &cflagSetter{Flag: syscall.PARODD},

	// Baud rates (ignore)
	ssh.TTY_OP_ISPEED: &nopSetter{},
	ssh.TTY_OP_OSPEED: &nopSetter{},
}

type nopSetter struct{}

type ccSetter struct {
	Character uint8
}

func (cc *ccSetter) Set(pty *os.File, termios *syscall.Termios, value uint32) error {
	termios.Cc[cc.Character] = byte(value)
	return SetAttr(pty, termios)
}

func (i *iflagSetter) Set(pty *os.File, termios *syscall.Termios, value uint32) error {
	if value == 0 {
		termios.Iflag &^= i.Flag
	} else {
		termios.Iflag |= i.Flag
	}
	return SetAttr(pty, termios)
}

func (l *lflagSetter) Set(pty *os.File, termios *syscall.Termios, value uint32) error {
	if value == 0 {
		termios.Lflag &^= l.Flag
	} else {
		termios.Lflag |= l.Flag
	}
	return SetAttr(pty, termios)
}

func (o *oflagSetter) Set(pty *os.File, termios *syscall.Termios, value uint32) error {
	if value == 0 {
		termios.Oflag &^= o.Flag
	} else {
		termios.Oflag |= o.Flag
	}

	return SetAttr(pty, termios)
}

func (c *cflagSetter) Set(pty *os.File, termios *syscall.Termios, value uint32) error {
	switch c.Flag {
	// CSIZE is a field
	case syscall.CS7, syscall.CS8:
		termios.Cflag &^= syscall.CSIZE
		termios.Cflag |= c.Flag
	default:
		if value == 0 {
			termios.Cflag &^= c.Flag
		} else {
			termios.Cflag |= c.Flag
		}
	}

	return SetAttr(pty, termios)
}

func (n *nopSetter) Set(pty *os.File, termios *syscall.Termios, value uint32) error {
	return nil
}
