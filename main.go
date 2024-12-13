package main

import (
	"fmt"
	"log"
	"log/slog"

	"github.com/bedrock-gophers/role/role"
	"github.com/df-mc/dragonfly/server"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/chat"
	"github.com/sandertv/gophertunnel/minecraft/text"
)

func main() {
	err := role.Load("assets/roles/", nil)
	if err != nil {
		panic(err)
	}
	fmt.Println(role.All())

	chat.Global.Subscribe(chat.StdoutSubscriber{})

	conf, err := server.DefaultConfig().Config(slog.Default())
	if err != nil {
		log.Fatalln(err)
	}

	srv := conf.New()
	srv.CloseOnProgramEnd()
	srv.Listen()

	for p := range srv.Accept() {
		p.Handle(&handler{})
	}
}

type handler struct {
	player.NopHandler
}

func (h *handler) HandleChat(ctx *player.Context, message *string) {
	ctx.Cancel()

	owner, ok := role.ByName("owner")
	if !ok {
		panic("role not found")
	}
	format := text.Colourf("<grey>[</grey>%s<grey>]</grey> %s<grey>:</grey> <white>%s</white>", owner.Coloured(owner.Name()), owner.Coloured(ctx.Val().Name()), *message)
	_, _ = chat.Global.WriteString(format)
}
