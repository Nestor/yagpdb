package botrest

import (
	"encoding/json"
	"github.com/Sirupsen/logrus"
	"github.com/jonas747/yagpdb/bot"
	"github.com/jonas747/yagpdb/common"
	"github.com/pkg/errors"
	"goji.io"
	"goji.io/pat"
	"net/http"
	"net/http/pprof"
	"strings"
)

var serverAddr = ":5002"

func StartServer() {
	muxer := goji.NewMux()
	muxer.Use(dropNonLocal)

	muxer.HandleFunc(pat.Get("/:guild/guild"), HandleGuild)
	muxer.HandleFunc(pat.Get("/:guild/botmember"), HandleBotMember)
	muxer.HandleFunc(pat.Get("/:guild/channelperms/:channel"), HandleChannelPermissions)
	muxer.HandleFunc(pat.Get("/ping"), HandlePing)

	// Debug stuff
	muxer.HandleFunc(pat.Get("/debug/pprof/*"), pprof.Index)
	muxer.HandleFunc(pat.Get("/debug/pprof"), pprof.Index)
	muxer.HandleFunc(pat.Get("/debug2/pproff/cmdline"), pprof.Cmdline)
	muxer.HandleFunc(pat.Get("/debug2/pproff/profile"), pprof.Profile)
	muxer.HandleFunc(pat.Get("/debug2/pproff/symbol"), pprof.Symbol)
	muxer.HandleFunc(pat.Get("/debug2/pproff/trace"), pprof.Trace)

	http.ListenAndServe(serverAddr, muxer)
}

func ServeJson(w http.ResponseWriter, r *http.Request, data interface{}) {
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		logrus.WithError(err).Error("Failed sending json")
	}
}

// Returns true if an error occured
func ServerError(w http.ResponseWriter, r *http.Request, err error) bool {
	if err == nil {
		return false
	}

	encodedErr, _ := json.Marshal(err.Error())

	w.WriteHeader(http.StatusInternalServerError)
	w.Write(encodedErr)
	return true
}

func dropNonLocal(inner http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Split(r.RemoteAddr, ":")[0] != "127.0.0.1" {
			logrus.Info("Dropped non local connection", r.RemoteAddr)
			return
		}

		inner.ServeHTTP(w, r)
	})
}

func HandleGuild(w http.ResponseWriter, r *http.Request) {
	gId := pat.Param(r, "guild")

	guild, err := bot.State.Guild(gId)
	if err != nil {
		ServerError(w, r, err)
		return
	}

	guild.Members = nil
	guild.Presences = nil
	guild.VoiceStates = nil

	ServeJson(w, r, guild)
}

func HandleBotMember(w http.ResponseWriter, r *http.Request) {
	gId := pat.Param(r, "guild")

	member, err := bot.State.GuildMember(gId, common.BotUser.ID)
	if err != nil {
		ServerError(w, r, err)
		return
	}

	ServeJson(w, r, member)
}

func HandleChannelPermissions(w http.ResponseWriter, r *http.Request) {
	cId := pat.Param(r, "channel")

	perms, err := bot.State.MemberPermissions(nil, cId, common.BotUser.ID)
	if err != nil {
		ServerError(w, r, errors.WithMessage(err, "Error calculating perms"))
		return
	}

	ServeJson(w, r, perms)
}

func HandlePing(w http.ResponseWriter, r *http.Request) {
	ServeJson(w, r, "pong")
}
