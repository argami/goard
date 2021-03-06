package main


import (
  "fmt"
  "io"
  "bytes"
  "log"
  "os"
  "path/filepath"
  "github.com/argami/goard/lxclib"

  
  "github.com/gin-gonic/gin"
  "github.com/lxc/lxd"
)

var config *lxd.Config


// Execute command, intercepts stdout and print info
func commandWrapper(c *gin.Context, command string, args []string) {
  old_stdout := os.Stdout // keep backup of the real stdout
  old_stderr := os.Stderr // keep backup of the real stdout
  r, w, _ := os.Pipe()
  os.Stdout = w
  os.Stderr = w

  err := lxclib.RunCommand(config, command, args)
  if err != nil {
    c.String(400, err.Error())
    c.Error(err)
  }

  outC := make(chan string)
  // copy the output in a separate goroutine so printing can't block indefinitely
  go func() {
      var buf bytes.Buffer
      io.Copy(&buf, r)
      outC <- buf.String()
  }()

  // back to normal state
  w.Close()
  os.Stdout = old_stdout // restoring the real stdout
  os.Stderr = old_stderr // restoring the real stdout
  out := <-outC

  c.String(200, out)
}

////////////////////////////////////////////////
// Web Handlers
///////////////////////////////////////////////

func webListContainers(c *gin.Context) {
  server := c.Param("remote")
  args := []string{server+":"}
  commandWrapper(c, "list", args) 
}

func webHelp(c *gin.Context) {
  command := c.Param("command")
  args := []string{}
  if command != "" {
    args = []string{command}
  }
  commandWrapper(c, "help", args) 
}

func webRemote(c *gin.Context) {
  remote := c.Param("remote")
  addr := c.Query("addr")
  password := c.Query("password")
      
  if password == "" { 
    c.String(400, "Need to define a remote server password")
    c.Error(fmt.Errorf("Need to define a remote server password"))
  }

  if addr == "" { 
    c.String(400, "Need to define a remote addr")
    c.Error(fmt.Errorf("Need to define a remote addr"))
  }
 
  args := []string{"add", remote, addr, "true", password, "true"}
  err := lxclib.RunCommand(config, "remote", args)
  if err != nil {
    c.String(400, err.Error())
    c.Error(err)
  }
  c.String(200, "Remote Added")
}

// URL: /snapshot/:remote/:container
func webSnapshot(c *gin.Context) {
  remote := c.Param("remote")
  container := c.Param("container")
  // snapname := c.Param("snapname")

  args := []string{remote + ":" + container}
  err := lxclib.RunCommand(config, "snapshot", args)
  if err != nil {
    c.String(400, err.Error())
    c.Error(err)
  }

  c.String(200, "Snapshot DONE")
}


// URL: /snapshot/:remote/:container
func webMove(c *gin.Context) {
  from := c.Param("from")
  to := c.Param("to")
  // snapname := c.Param("snapname")

  args := []string{from, to}
  err := lxclib.RunCommand(config, "move", args)
  if err != nil {
    c.String(400, err.Error())
    c.Error(err)
  }

  c.String(200, "Move DONE")
}

////////////////////////////////////////////////
// Main
///////////////////////////////////////////////

func main() {
    dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
    if err != nil {
            log.Fatal(err)
    }
    fmt.Println(dir)
    fmt.Println(os.Args[0])
    lxd.ConfigDir = dir 
    config, _ = lxd.LoadConfig()
    
    r := gin.Default()
    r.GET("/help/*command", webHelp)
    r.GET("/list/:remote", webListContainers)
    r.GET("/remote/add/:remote", webRemote)
    r.GET("/snapshot/:remote/:container", webSnapshot)
    r.GET("/move/:from/:to/", webMove)
    r.Run(":8080") // listen and serve on 0.0.0.0:8080
  }
