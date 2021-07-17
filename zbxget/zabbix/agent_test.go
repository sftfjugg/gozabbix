package zabbix_test

import (
	"fmt"
	"log"
	"net"
	"testing"
	"time"

	"github.com/sftfjugg/gozabbix/zbxget/zabbix"
)

var (
	done    chan (bool)
	timeout = 3 * time.Second
)

func TestAgent(t *testing.T) {
	done = make(chan bool)
	go zabbix.RunAgent("localhost", func(key string) (string, error) {
		switch key {
		case "agent.ping":
			log.Println("key", key)
			return "1", nil
		case "agent.uptime":
			log.Println("key", key)
			return "123", nil
		case "timeout":
			log.Println("key", key, "sleeping...")
			time.Sleep(timeout + time.Second)
			log.Println("wake up. response ok!")
			return "ok", nil
		case "shutdown":
			done <- true
			return "", nil
		default:
			return "", fmt.Errorf("not supported")
		}
	})
	time.Sleep(1 * time.Second)
}

func TestAgentCannotConnect(t *testing.T) {
	value, err := zabbix.Get("localhost:10049", "agent.ping", timeout)
	if err == nil {
		t.Errorf("agent is not runnig, but not error value:", value)
	}
}

func TestAgentGetPing(t *testing.T) {
	value, err := zabbix.Get("localhost", "agent.ping", timeout)
	if err != nil {
		t.Error("get agent.ping failed", err)
	}
	if value != "1" {
		t.Error("agent.ping value expected: 1, got:", value)
	}
}

func TestAgentGetUptime(t *testing.T) {
	value, err := zabbix.Get("localhost", "agent.uptime", timeout)
	if err != nil {
		t.Error("get agent.ping failed", err)
	}
	if value != "123" {
		t.Error("agent.uptime value expected: 123, got:", value)
	}
}

func TestAgentGetError(t *testing.T) {
	value, err := zabbix.Get("localhost", "xxx", timeout)
	if err != nil {
		t.Error("xxx failed", err)
	}
	if value != zabbix.ErrorMessage {
		t.Error("xxx value expected: ", zabbix.ErrorMessage, "got:", value)
	}
}

func TestAgentGetTimeout(t *testing.T) {
	_, err := zabbix.Get("localhost", "timeout", timeout)
	if err == nil {
		t.Error("timeout must be timeouted.", err)
	}
	if _err := err.(*net.OpError); !_err.Timeout() {
		t.Error("err expected i/o timeout. got:", err)
	}
	log.Println("client timeout")
}

func TestAgentShutdown(t *testing.T) {
	zabbix.Get("localhost", "shutdown", timeout)
	<-done
}
