---
title: "Building a Personal Activity Tracker"
date: 2021-04-04
draft: false
---

> **Disclaimer**: This project/post is inspired on [Building an activity tracker with Go, Grafana, and InfluxDB](http://lucapette.me/building-an-activity-tracker-with-go-grafana-and-influxdb) by [lucapette](https://twitter.com/lucapette), so there're many common points and strategies.

![the dashboard](/img/building-a-personal-activity-tracker/productivity_dash.png)

If you're reading this in 2021 you're living the [Coronavirus Pandemic Era](https://www.who.int/health-topics/coronavirus) and you're most likely working from home.
Well, I'm working from home too with my fiance and my daughter (and two cats).
I'm a software engineer who works in a big retail company and this kind of job is very intense.
To archive a good work/life balance, I try to do a great focused job on my working hours, but when it's time to rest I stop any job or study activity to give total attention to my family.
I don't know how you handle it but for my is a bit tricky to get focused working from home with my family, so I'm trying so hard to get rid of procrastination.
To help me with this challenge I decided to build a simple activity tracker that monitoring my activities and give them a score based on some simple rules:

* **Work**: 1 point
* **Social**: 0 point
* **Game**: -1 point

Don't get me wrong, there's no problem with gaming and using social apps, but when I'm working I need to use my time with wisdom.

## How I built this?
The first POC was built for Ubuntu environment so there're some components to get all the project up and running:

* InfluxDB to keep the time series data.
* A Shell script to get the current X window activity and send it to API.
* A Go API to receive activity events, treat the data, and save on InfluxDB.
* Grafana to show activities dashboard.

Just for fun, I deployed every component (except the shell script) on an old Raspberry Pi 1 running [DietPi](https://dietpi.com/) distro.

####  Maybe you're thinking, "Why"?
It's ok if you're thinking about my decisions but I'll try to explain some points.

##### Why deploy it on a Raspberry Pi?
The first (and maybe the main) reason is I had an old Raspberry Pi 1 saved for nothing so I decided to use it.
The second reason is I'm working mostly with a Macbook but I study with an Ubuntu and I tried to create a solution that works in both scenarios.

##### Why Golang?
It's simple, as long I decided to ship the solution on a Raspberry Pi, Golang fits very well to produce a binary that runs on the Raspberry Pi without any virtual machine or system dependency.
Also, the Golang compiler is awesome, I didn't have any problem compiling to Raspberry Pi (and I love to write Golang code).

### The Shell Script
Let's talk about the shell script.
As I said before, I'm using Ubuntu with Gnome, so I can use the [xprop](https://linux.die.net/man/1/xprop) utility to get some information about the focused window of the X server.
The usage of _xprop_ utility is simple, f.ex:

```
$ xprop -root _NET_ACTIVE_WINDOW
_NET_ACTIVE_WINDOW(WINDOW): window id # 0x3c0000a
```

The important thing on that output is the window id **0x3c0000a**.
With that id we can dig through the window information, f.ex:

```
$ xprop -id 0x3c0000a WM_CLASS
WM_CLASS(STRING) = "gnome-terminal-server", "Gnome-terminal"
```

As you can see, now we have the name of the focused window, but we still need to know the title of the window to specify what we're doing in that application.
In the previous command we asked for _the class_ of the window _0x3c0000a_ however we can ask for the _name_ of the window (or title), f.ex:

```
$ xprop -id 0x3c0000a WM_NAME
WM_NAME(STRING) = "default - zsh"
```

I'm using Tmux with ZSH while I write this blog post, so the **default** is the name of my Tmux session, and the **zsh** is the currently running application on the _gnome-terminal-server_.
Now that we have all the information we need, we can send it somewhere with _curl_.

```
$ curl -v -H "Content-Type: application/json" -d \
    "{\"class\": \"$window_class\", \"title\": \"$window_title\"}" \
    $ENDPOINT
```

### The Go API
This is our _business logic layer_, in this API we interpret the data from the collector shell script and persist it into InfluxDB.
But, don't be scared, this is a very simple (maybe a naive) API made only with the Golang standard library.
There're three main functions on that API to solve our problem:

1. A HTTP handler to receive the collector's requests.
2. A transformation layer to convert incoming data into the score rules.
3. A InfluxDB _write-only_ client.

A simplified version of the handle contains less than 15 LOC:

```golang
func (s *Server) track(w http.ResponseWriter, r *http.Request) {
	var d requestData
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
        // error handling for a "Bad Request"
	}
	if err := s.client.Write(d.Class, d.Title); err != nil {
        // error handling data processing and saving data on influxdb
	}
	w.WriteHeader(http.StatusNoContent)
}
```
The `client` is our InfluxDB client.
Before we getting through the InfluxDB client let's take a look at the transformation layer (that's the ugly layer of this API).

```golang
func convertToMetric(class, title string) metric {
	switch class {
	case "gnome-terminal-server":
		return metric{
			category: WORK_CATEGORY,
			app:      convertTerminalAppName(title),
		}
	case "Navigator":
		return metric{
			category: convertBrowserCategory(title),
			app:      "FireFox",
		}
	case "telegram":
		return metric{
			category: SOCIAL_CATEGORY,
			app:      class,
		}
	case "Steam":
		return metric{
			category: GAME_CATEGORY,
			app:      "Steam",
		}
	default:
		return metric{category: UNKNOWN_CATEGORY, app: title}
	}
}
```

Every rule of the project is described above, for example, if the collector sends a request like this:
```json
{"class": "gnome-terminal-server", "title": "default - vim"}
```
The converter will return a metric with the `WORK_CATEGORY`:
```golang
metric{
    category: category{name: "work", score: 1},
    app     : "vim",
}
```
There're other rules like verify if the current tab focused on the browser is a social site or not but, you get the idea.
So, back to the InfluxDB client, actually the InfluxDB has a friendly REST API and because of that I decided to avoid big dependencies and solve the persistence layer with a simple _HTTP POST_:

```golang
func (c *Client) Write(class, title string) error {
	resp, err := http.Post(
		fmt.Sprintf("%s/write?db=%s", c.endpoint, c.database),
		"",
		buildPayload(class, title),
	)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusNoContent {
        // boring error handling
	}
	return nil
}
```
Yeah, simple as pie.
If you're curious about the payload, here's an example:

```ini
productivity,category="work",app="vim" value=1
```
The last thing I want to share about the Go API is the compilation step.
I deployed that API on a Raspberry Pi 1 and according to [Golang Wiki](https://github.com/golang/go/wiki/GoArm), to compile to _ARMv6_ CPU architecture we need to pass the environment variable `GOARM=6` on the build command (and the env `GOARCH=arm` too).

```
$ GOARCH=arm GOARM=6 go build
```

### InfluxDB

There're no tricks to get InfluxDB up and running on the Raspberry Pi, the straightforward `apt install influxdb` solves the problem.
After the installation process, I created the database and set the _retention policy_ to keep only one week of data (it's enough for what I'm looking for).

```
$ influx
Connected to http://localhost:8086 version 1.8.4
InfluxDB shell version: 1.8.4
> CREATE DATABASE productivity
> USE productivity
> CREATE RETENTION POLICY "1week" on "productivity" DURATION 7d REPLICATION 1
```

### Grafana

However to install Grafana on the Raspberry Pi 1 there's a trick, the default version of Grafana's Debian package is built for the _ARMv7_ architecture (its means that a simple `apt install grafana` doesn't work), but the DietPi has the `dietpi-software` application that enables you to select an optimized installation of some software for the Raspberry Pi with DietPi OS.
So I just installed Grafana from that and everything works well.
As long the Grafana and the InfluxDB are up and running I created the InfluxDB data source on the Grafana and start "to draw" the dashboard.

#### All Activities Panel

For the _All Activities Panel_ I used a graph visualization with bars as display option and the following query:

```
SELECT count("value") as minutes FROM "productivity" WHERE $timeFilter GROUP BY time(1h),category
```

#### Score Panel
The _Score_ panel is a gauger that calculates the _mean_ of the following query:
```
SELECT mean("value") * 100 FROM "productivity" WHERE $timeFilter GROUP BY time($interval) fill(null)
```
I multiplied to 100 to get the gauger working like a "percentage".

#### App Details Panel
_App Details_ panel shows the name of apps with the time spending with them. It's a table panel and I'm using the following query:
```
SELECT count("value") FROM "productivity" WHERE $timeFilter GROUP BY time($interval), app fill(null)
```

#### Top Unknown
And the last one is the _top unknown_ panel, a panel to show the apps that are still uncategorized

```
SELECT count(value) FROM "productivity" WHERE $timeFilter and category = '"unknown"' GROUP BY app
```

### Turning the Go API into a Linux service
To keep the Go API up and running I created a systemd service.
To do that I created a file named `productivity.service` on directory `/etc/systemd/system/` with this content:

```ini
[Unit]
Description=The Productivity Tracker API

[Service]
User=tracker
WorkingDirectory=/home/productivity/app
ExecStart=/home/tracker/app/productivity
Restart=always

[Install]
WantedBy=multi-user.target
```

After that, I added the service and started it

```bash
$ systemctl enable productivity.service
$ systemctl start productivity.service
```

## The Source Code
I pushed everything of this POC to a Github repo but if you want to use it, you'll need to do some changes (this project was made to help me with my productivity and to have some fun).
I know it's weird but I named this project as [Floki](https://vikings.fandom.com/wiki/Floki), I thought Floki is a better name than _productivity_ or _tracker_.
