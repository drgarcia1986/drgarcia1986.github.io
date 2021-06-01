---
title: "Changelogger"
date: 2021-06-01
draft: false
---
I started a project called `changelogger`.
This project is a very simplified version of [Towncrier](https://github.com/twisted/towncrier) but written in Go.

I can already hear you say: **Why?**.
Well, Golang isn't the main goal here, but the fact that I can distribute this tool as a binary without care about the developer environment.
There's no problem with Python (to tell the truth, Python has been paying my bills for ten years) but, I wouldn't like to force a _Golang/JavaScript/Java/[put here your favorite lang]_ dev to setup a python environment only to use a simple tool.

I used towncrier in my last python projects and I just love how the tool helps me to keep my changelog updated without git conflicts, so I want to continue to use it (or some similarly tool).

That's how changelogger was born, with only limited features that I'm used to using in towncrier.

Check [the repository](https://github.com/drgarcia1986/changelogger), get the binary on [release page](https://github.com/drgarcia1986/changelogger/releases/tag/v0.0.1) or install from the source with `go get github.com/drgarcia1986/changelogger`.
