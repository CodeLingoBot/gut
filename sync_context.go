package main

import (
    "errors"
    "fmt"
    "path"
    "regexp"
    "github.com/tillberg/bismuth"
)

type SyncContext struct {
    *bismuth.ExecContext
    syncPath string
}

func NewSyncContext() *SyncContext {
    ctx := &SyncContext{}
    ctx.ExecContext = &bismuth.ExecContext{}
    ctx.Init()
    return ctx
}

var remotePathRegexp = regexp.MustCompile("^((([^@]+)@)?([^:]+):)?(.+)$")
func (ctx *SyncContext) ParseSyncPath(path string) error {
    parts := remotePathRegexp.FindStringSubmatch(path)
    if len(parts) == 0 {
        return errors.New(fmt.Sprintf("Could not parse remote path: [%s]\n", path))
    }
    isRemote := len(parts[1]) > 0
    if isRemote {
        if len(parts[3]) > 0 {
            ctx.SetUsername(parts[3])
        }
        ctx.SetHostname(parts[4])
    }
    ctx.syncPath = parts[5]
    return nil
}

func (ctx *SyncContext) AbsSyncPath() string {
    return ctx.AbsPath(ctx.syncPath)
}

func (ctx *SyncContext) String() string {
    if ctx.Hostname() != "" {
        return fmt.Sprintf("{SyncContext %s@%s:%s}", ctx.Username(), ctx.Hostname(), ctx.syncPath)
    }
    return fmt.Sprintf("{SyncContext local %s}", ctx.syncPath)
}

func (ctx *SyncContext) PathAnsi(p string) string {
    if !ctx.IsLocal() {
        return fmt.Sprintf(ctx.Logger().Colorify("%s@(dim::)@(path:%s)"), ctx.NameAnsi(), p)
    }
    return fmt.Sprintf(ctx.Logger().Colorify("@(path:%s)"), p)
}

func (ctx *SyncContext) SyncPathAnsi() string {
    return ctx.PathAnsi(ctx.syncPath)
}

func (ctx *SyncContext) Mkdirp(p string) (err error) {
    if ctx.IsWindows() {
        return errors.New("Not implemented")
    }
    _, err = ctx.Output("mkdir", "-p", ctx.AbsPath(p))
    return err
}

func (ctx *SyncContext) GutExe() string {
    return ctx.AbsPath(GutExePath)
}

func (ctx *SyncContext) gutArgs(otherArgs ...string) []string {
    args := []string{}
    args = append(args, ctx.GutExe())
    return append(args, otherArgs...)
}

func (ctx *SyncContext) GutOutput(args ...string) (string, error) {
    return ctx.OutputCwd(ctx.AbsSyncPath(), ctx.gutArgs(args...)...)
}

func (ctx *SyncContext) GutQuote(suffix string, args ...string) error {
    return ctx.QuoteCwd(suffix, ctx.AbsSyncPath(), ctx.gutArgs(args...)...)
}

func (ctx *SyncContext) getPidfilePath(name string) string {
    return ctx.AbsPath(path.Join(PidfilesPath, name + ".pid"))
}

func (ctx *SyncContext) SaveDaemonPid(name string, pid int) (err error) {
    err = ctx.Mkdirp(PidfilesPath)
    if err != nil { ctx.Logger().Bail(err) }
    return ctx.WriteFile(ctx.getPidfilePath(name), []byte(fmt.Sprintf("%d", pid)))
}
