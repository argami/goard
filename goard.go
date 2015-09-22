package main


import (
  "fmt"
  "os"
  "path/filepath"
  "log"
  "net"
  "net/url"
  "strings"
  
  "github.com/gin-gonic/gin"
  "github.com/lxc/lxd"
  "github.com/lxc/lxd/shared"
)

var config *lxd.Config

// URL: /:server/:container/snapshot/:snapname
func doSnapshot(c *gin.Context) {
  server := c.Param("server")
  container := c.Param("container")
  snapname := c.Param("snapname")

  remote := config.ParseRemote(server)
  client, err := lxd.NewClient(config, remote)
  if err != nil {
   c.Error(err) //return err
  }

	// we don't allow '/' in snapshot names
	if shared.IsSnapshot(snapname) {
    c.Error(fmt.Errorf("'/' not allowed in snapshot name"))
	}

	resp, err := client.Snapshot(container, snapname, false)
	if err != nil {
	  c.Error(err)
  }

  err = client.WaitForSuccess(resp.Operation)
  if err != nil {
    c.Error(err)
  }

  c.String(200, "Snapshot DONE")
}


func listContainers(c *gin.Context) {
      server := c.Param("server")

      remote := config.ParseRemote(server)
	    client, err := lxd.NewClient(config, remote)
	    if err != nil {
       c.Error(err) //return err
	    }

      ctslist, err := client.ListContainers()
      if err != nil {
        c.Error(err) //return err
      }

      for _, cinfo := range ctslist {
        c.String(200, cinfo.State.Name)
      }
}


func addRemote(c *gin.Context) {
      server := c.Param("server")
      addr := c.Query("addr")
      password := c.Query("password")
      
      if password == "" { 
        c.Error(fmt.Errorf("Need to define a remote server password"))
		  }

      if addr == "" { 
        c.Error(fmt.Errorf("Need to define a remote addr"))
		  }
      
      if rc, ok := config.Remotes[server]; ok {
        c.Error(fmt.Errorf("remote %s exists as <%s>", server, rc.Addr))
		  }
      
      err := configRemote(config, server, addr, true, password, false)
	    if err != nil {
        c.Error(err) //return err
	    }
      lxd.SaveConfig(config)

      c.String(200, "Remote Added")
}


func configRemote(config *lxd.Config, server string, addr string, acceptCert bool, password string, public bool) error {
	var r_scheme string
	var r_host string
	var r_port string

	/* Complex remote URL parsing */
	remote_url, err := url.Parse(addr)
	if err != nil {
		return err
	}

	if remote_url.Scheme != "" {
		if remote_url.Scheme != "unix" && remote_url.Scheme != "https" {
			r_scheme = "https"
		} else {
			r_scheme = remote_url.Scheme
		}
	} else if addr[0] == '/' {
		r_scheme = "unix"
	} else {
		if !shared.PathExists(addr) {
			r_scheme = "https"
		} else {
			r_scheme = "unix"
		}
	}

	if remote_url.Host != "" {
		r_host = remote_url.Host
	} else {
		r_host = addr
	}

	host, port, err := net.SplitHostPort(r_host)
	if err == nil {
		r_host = host
		r_port = port
	} else {
		r_port = shared.DefaultPort
	}

	if r_scheme == "unix" {
		if addr[0:5] == "unix:" {
			if addr[0:7] == "unix://" {
				r_host = addr[8:]
			} else {
				r_host = addr[6:]
			}
		}
		r_port = ""
	}

	if strings.Contains(r_host, ":") && !strings.HasPrefix(r_host, "[") {
		r_host = fmt.Sprintf("[%s]", r_host)
	}

	if r_port != "" {
		addr = r_scheme + "://" + r_host + ":" + r_port
	} else {
		addr = r_scheme + "://" + r_host
	}

	if config.Remotes == nil {
		config.Remotes = make(map[string]lxd.RemoteConfig)
	}

	/* Actually add the remote */
	config.Remotes[server] = lxd.RemoteConfig{Addr: addr, Public: public}

	remote := config.ParseRemote(server)
	c, err := lxd.NewClient(config, remote)
	if err != nil {
		return err
	}

	if len(addr) > 5 && addr[0:5] == "unix:" {
		// NewClient succeeded so there was a lxd there (we fingered
		// it) so just accept it
		return nil
	}

	err = c.UserAuthServerCert(host, acceptCert)
	if err != nil {
		return err
	}

	if public {
		if err := c.Finger(); err != nil {
			return err
		}

		return nil
	}

	if c.AmTrusted() {
		// server already has our cert, so we're done
		return nil
	}
  if password == ""  {
		return fmt.Errorf("Password not provided")
	}

	err = c.AddMyCertToServer(password)
	if err != nil {
		return err
	}

	if !c.AmTrusted() {
		return fmt.Errorf("Server doesn't trust us after adding our cert")
	}

	fmt.Println("Client certificate stored at server: ", server)
	return nil
}


func main() {
    dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
    if err != nil {
            log.Fatal(err)
    }
    lxd.ConfigDir = dir 
    config, _ = lxd.LoadConfig()
    
    r := gin.Default()
    r.GET("/:server/snapshot/:container/:snapname", doSnapshot)
    r.GET("/:server/list_containers", listContainers)
    r.GET("/:server/remote/add/", addRemote)
    r.Run(":8080") // listen and serve on 0.0.0.0:8080
}
