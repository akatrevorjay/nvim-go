package commands

import (
	"fmt"
	"go/build"
	"time"

	"nvim-go/config"
	"nvim-go/context"
	"nvim-go/nvim"
	"nvim-go/nvim/profile"

	"github.com/garyburd/neovim-go/vim"
	"github.com/garyburd/neovim-go/vim/plugin"
	"golang.org/x/tools/refactor/rename"
)

func init() {
	plugin.HandleCommand("Gorename",
		&plugin.CommandOptions{
			NArgs: "?", Eval: "[expand('%:p:h'), expand('%:p'), line2byte(line('.'))+(col('.')-2), expand('<cword>')]"},
		cmdRename)
}

type cmdRenameEval struct {
	Dir    string `msgpack:",array"`
	File   string
	Offset int
	From   string
}

func cmdRename(v *vim.Vim, args []string, eval *cmdRenameEval) {
	go Rename(v, args, eval)
}

// Rename rename the current cursor word use golang.org/x/tools/refactor/rename.
func Rename(v *vim.Vim, args []string, eval *cmdRenameEval) error {
	defer profile.Start(time.Now(), "GoRename")
	var ctxt = context.Build{}
	defer ctxt.SetContext(eval.Dir)()

	var (
		b vim.Buffer
		w vim.Window
	)
	p := v.NewPipeline()
	p.CurrentBuffer(&b)
	p.CurrentWindow(&w)
	if err := p.Wait(); err != nil {
		return err
	}

	offset := fmt.Sprintf("%s:#%d", eval.File, eval.Offset)

	var to string
	if len(args) > 0 {
		to = args[0]
	} else {
		askMessage := fmt.Sprintf("%s: Rename '%s' to: ", "GoRename", eval.From)
		var toResult interface{}
		if config.RenamePrefill {
			p.Call("input", &toResult, askMessage, eval.From)
			if err := p.Wait(); err != nil {
				return nvim.EchohlErr(v, "GoRename", err)
			}
		} else {
			p.Call("input", &toResult, askMessage)
			if err := p.Wait(); err != nil {
				return nvim.EchohlErr(v, "GoRename", err)
			}
		}
		if toResult.(string) != "" {
			to = toResult.(string)
		}
	}

	prefix := "GoRename"
	nvim.EchoProgress(v, prefix, "Renaming", eval.From, to)
	defer nvim.EchoSuccess(v, prefix)

	if err := rename.Main(&build.Default, offset, "", to); err != nil {
		if err != rename.ConflictError {
			return err
		}
	}
	p.Command("silent! edit!")

	return p.Wait()
}
