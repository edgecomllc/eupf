#!/bin/bash

set -eu

iperf_source_ip=$(ip a show dev uesimtun0 | grep inet | awk '{print $2}' | awk -F'/' '{print $1}')
iperf_result_file="/opt/iperf/results/${MSISDN}.txt"

iperf3 -c ${IPERF_HOST} -p ${IPERF_PORT} -t ${IPERF_TIME} -B ${iperf_source_ip} | tee ${iperf_result_file}
