// Copyright 2018 gf Author(https://github.com/gogf/gf). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/gogf/gf.

package ghttp_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/net/ghttp"
	"github.com/gogf/gf/test/gtest"
)

func Test_Cookie(t *testing.T) {
	p := ports.PopRand()
	s := g.Server(p)
	s.BindHandler("/set", func(r *ghttp.Request) {
		r.Cookie.Set(r.GetString("k"), r.GetString("v"))
	})
	s.BindHandler("/get", func(r *ghttp.Request) {
		r.Response.Write(r.Cookie.Get(r.GetString("k")))
	})
	s.BindHandler("/remove", func(r *ghttp.Request) {
		r.Cookie.Remove(r.GetString("k"))
	})
	s.SetPort(p)
	s.SetDumpRouteMap(false)
	s.Start()
	defer s.Shutdown()

	// 等待启动完成
	time.Sleep(200 * time.Millisecond)
	gtest.Case(t, func() {
		client := ghttp.NewClient()
		client.SetBrowserMode(true)
		client.SetPrefix(fmt.Sprintf("http://127.0.0.1:%d", p))
		r1, e1 := client.Get("/set?k=key1&v=100")
		if r1 != nil {
			defer r1.Close()
		}
		gtest.Assert(e1, nil)
		gtest.Assert(r1.ReadAllString(), "")

		gtest.Assert(client.GetContent("/set?k=key2&v=200"), "")

		gtest.Assert(client.GetContent("/get?k=key1"), "100")
		gtest.Assert(client.GetContent("/get?k=key2"), "200")
		gtest.Assert(client.GetContent("/get?k=key3"), "")
		gtest.Assert(client.GetContent("/remove?k=key1"), "")
		gtest.Assert(client.GetContent("/remove?k=key3"), "")
		gtest.Assert(client.GetContent("/remove?k=key4"), "")
		gtest.Assert(client.GetContent("/get?k=key1"), "")
		gtest.Assert(client.GetContent("/get?k=key2"), "200")
	})
}
