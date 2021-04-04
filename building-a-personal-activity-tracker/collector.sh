#!/bin/bash

get_current_focused_window_id() {
    echo $(xprop -root _NET_ACTIVE_WINDOW | awk '{print $NF}')
}

get_xprop_info() {
    WINDOW_ID=$1
    INFO_TYPE=$2
    PATTERN=$3

    result=$(xprop -id $WINDOW_ID $INFO_TYPE)
    if [[ "$result" =~ $PATTERN ]]; then
        echo ${BASH_REMATCH[1]}
    else
        echo $result
    fi
}

get_window_class() {
    get_xprop_info $1 "WM_CLASS" '"([^,]+)"'
}

get_window_title() {
    get_xprop_info $1 "WM_NAME" '"(.*)"'
}

ENDPOINT=$1

while true; do
    current_window_id=$(get_current_focused_window_id)
    window_class=$(get_window_class $current_window_id)
    window_title=$(get_window_title $current_window_id)

    if [[ -z "$window_class" || -z "$window_title" ]]; then
        continue
    fi

    curl -v -H "Content-Type: application/json" -d "{\"class\": \"$window_class\", \"title\": \"$window_title\"}" $ENDPOINT
    sleep 2;
done
