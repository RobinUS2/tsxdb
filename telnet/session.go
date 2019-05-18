package telnet

import (
	"fmt"
	"github.com/RobinUS2/tsxdb/client"
	"github.com/pkg/errors"
	"github.com/reiver/go-oi"
	tel "github.com/reiver/go-telnet"
	"log"
	"strconv"
	"strings"
)

const successMessage = "+OK"
const errorMessage = "-ERR"
const redisAddToSortedSetCommand = "ZADD"              // ZADD key [NX|XX] [CH] [INCR] score member [score member ...] https://redis.io/commands/zadd
const redisRangeFromSortedSetCommand = "ZRANGEBYSCORE" // ZRANGEBYSCORE key min max [WITHSCORES] [LIMIT offset count] https://redis.io/commands/zrangebyscore

type Mode string

const ModePlain Mode = "PLAIN" // auth
const ModeRedis Mode = "REDIS" // *x $y zzzz => *1 $4 auth

type Session struct {
	instance      *Instance
	writer        tel.Writer
	authenticated bool
	mode          Mode
	client        *client.Instance
}

func (session *Session) SetMode(mode Mode) {
	log.Printf("session in mode %s", mode)
	session.mode = mode
}

func (session *Session) Handle(typedLine InputLine) error {
	line := strings.TrimSpace(string(typedLine))
	if len(line) < 1 {
		return nil
	}
	log.Printf("telnet rcv %s", line)

	// tokens
	tokens := strings.Split(line, " ")
	command := strings.ToUpper(tokens[0])

	// auth
	if command == "AUTH" {
		if len(tokens) < 2 {
			return session.WriteErrMessage(errors.New("missing auth token"))
		}

		// first check local
		token := tokens[1]
		if token != session.instance.opts.AuthToken {
			return session.WriteErrMessage(errors.New("invalid auth token"))
		}

		// real remote auth
		clientOpts := client.NewOpts()
		clientOpts.AuthToken = token
		clientOpts.ListenHost = session.instance.opts.ServerHost
		clientOpts.ListenPort = session.instance.opts.ServerPort
		session.client = client.New(clientOpts)

		// this verifies auth with the server
		_, err := session.client.GetConnection()
		if err != nil {
			return errors.Wrap(err, "fail to get connection")
		}

		// good
		session.authenticated = true

		// yeah
		if err := session.Write(successMessage); err != nil {
			return err
		}

		// done
		return nil
	}

	// authenticated?
	if !session.authenticated {
		return session.WriteErrMessage(errors.New("not authenticated"))
	}

	// echo back
	if command == "ECHO" {
		// echo, as error, since redis else can't handle it
		return session.WriteErrMessage(errors.New(strings.Replace(line, "ECHO ", "", 1)))
	} else if command == redisAddToSortedSetCommand {
		// add to serie
		// zadd mySeries 23456789 10.0
		//log.Printf("zadd %+v", tokens)
		seriesName := tokens[1]
		ts, _ := strconv.ParseUint(tokens[2], 10, 64)
		val, _ := strconv.ParseFloat(tokens[3], 64)
		series := session.client.Series(seriesName)
		res := series.Write(ts, val)
		if res.Error != nil {
			return res.Error
		}
		return session.Write(successMessage)
	} else if command == redisRangeFromSortedSetCommand {
		// get from serie
		// ZRANGEBYSCORE abc -inf +inf should return all values
		seriesName := tokens[1]
		from, _ := strconv.ParseUint(tokens[2], 10, 64)
		to, _ := strconv.ParseUint(tokens[3], 10, 64)
		series := session.client.Series(seriesName)
		qb := series.QueryBuilder()
		qb.From(from)
		qb.To(to)
		res := qb.Execute()
		if res.Error != nil {
			return res.Error
		}
		log.Printf("qres %v", res)
		return session.Write(successMessage)
	} else {
		// command not found
		return session.WriteErrMessage(errors.New(fmt.Sprintf("command %s not found", command)))
	}
}

func (session *Session) SetWriter(writer tel.Writer) {
	session.writer = writer
}

func (session *Session) WriteErrMessage(err error) error {
	return session.Write(errorMessage + " " + err.Error())
}

func (session *Session) Write(s string) error {
	s = strings.TrimRight(s, "\r\n")
	if !strings.HasSuffix(s, "\r\n") {
		s = s + "\r\n"
	}
	log.Printf("telnet send %s", s)
	b := []byte(s)
	if nWritten, err := oi.LongWrite(session.writer, b); err != nil || int64(len(b)) != nWritten {
		return err
	}
	return nil
}

func NewSession(instance *Instance) *Session {
	return &Session{
		instance: instance,
		mode:     ModePlain,
	}
}

type InputLine string
