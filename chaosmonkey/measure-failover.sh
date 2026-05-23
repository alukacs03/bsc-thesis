#!/bin/bash
# Failover mérés:
#  - probe: SSH-zik egy node-ra, onnan ping-el egy célpont overlay IP-t
#  - inject: SSH-zik egy másik node-ra, ott blokkol egy interfészt iptables-szel
#  - mérés: első csomagvesztés és első recovery időpontja, csomagvesztés száma

set -e

PROBE_HOST="${1:-192.168.1.24}"        # ahonnan ping-elünk (testw02)
TARGET_IP="${2:-10.255.0.5}"           # cél overlay IP (testw01 loopback)
FAULT_HOST="${3:-192.168.1.24}"        # ahol az interfészt blokkoljuk (testw02 saját maga)
FAULT_INTERFACE="${4:-wg-hub1}"        # blokkolandó interfész
FAULT_DURATION="${5:-30}"              # hibaidő mp-ben
PING_INTERVAL="${6:-0.2}"              # ping intervallum mp-ben

EXP_ID="exp-$(date +%Y%m%d-%H%M%S)"
PING_LOG="/tmp/${EXP_ID}-ping.log"

echo "=== Failover mérés ==="
echo "Probe:      ${PROBE_HOST}"
echo "Target:     ${TARGET_IP}"
echo "Fault:      ${FAULT_HOST} -> ${FAULT_INTERFACE} drop ${FAULT_DURATION}s"
echo

echo "Ping indítása ${PROBE_HOST}-on a ${TARGET_IP}-re..."
ssh -o ServerAliveInterval=10 alukacs@${PROBE_HOST} \
  "ping -i ${PING_INTERVAL} -W 1 ${TARGET_IP}" > "${PING_LOG}" 2>&1 &
PING_PID=$!

sleep 3

START_TS=$(date +%s.%N)
echo "[$(date +%T.%3N)] Hiba injektálása..."
ssh alukacs@${FAULT_HOST} "sudo iptables -I INPUT -i ${FAULT_INTERFACE} -j DROP -m comment --comment ${EXP_ID}; sudo iptables -I OUTPUT -o ${FAULT_INTERFACE} -j DROP -m comment --comment ${EXP_ID}"

sleep ${FAULT_DURATION}

END_TS=$(date +%s.%N)
echo "[$(date +%T.%3N)] Hiba visszavonása..."
ssh alukacs@${FAULT_HOST} "sudo iptables -D INPUT -i ${FAULT_INTERFACE} -j DROP -m comment --comment ${EXP_ID}; sudo iptables -D OUTPUT -o ${FAULT_INTERFACE} -j DROP -m comment --comment ${EXP_ID}"

sleep 5

kill ${PING_PID} 2>/dev/null || true
wait ${PING_PID} 2>/dev/null || true

echo
echo "=== Eredmények ==="
TOTAL=$(grep -c "icmp_seq=" "${PING_LOG}" || echo 0)
LOST=$(grep -c "no answer" "${PING_LOG}" || echo 0)
RECEIVED=$(grep -c "bytes from" "${PING_LOG}" || echo 0)

echo "Összes ping próba (recv-ed):     ${RECEIVED}"
echo "Időtartam (hiba+recovery):       $(echo "${END_TS} - ${START_TS} + 5" | bc) mp"

# parse ping output - timestamps are not in default ping output, so we use sequence numbers
# Find first gap in icmp_seq numbers
FIRST_LOSS=$(awk -F'icmp_seq=' '
  /bytes from/ {
    n = $2 + 0
    if (last && n > last + 1) {
      print last + 1
      exit
    }
    last = n
  }' "${PING_LOG}")

# Find recovery: the seq number after the gap that succeeded
RECOVERY=$(awk -F'icmp_seq=' '
  /bytes from/ {
    n = $2 + 0
    if (last && n > last + 1) {
      print n
      exit
    }
    last = n
  }' "${PING_LOG}")

if [ -n "${FIRST_LOSS}" ] && [ -n "${RECOVERY}" ]; then
  GAP=$(( RECOVERY - FIRST_LOSS ))
  FAILOVER_TIME=$(echo "${GAP} * ${PING_INTERVAL}" | bc)
  echo "Első csomagvesztés icmp_seq:     ${FIRST_LOSS}"
  echo "Első sikeres ping recovery után: ${RECOVERY}"
  echo "Elveszett csomagok száma:        ${GAP}"
  echo "Failover idő:                    ~${FAILOVER_TIME} mp"
else
  echo "FIGYELEM: nincs detektálható csomagvesztés (lehet, hogy a hiba nem hatott a routingra)"
fi

echo
echo "Ping log: ${PING_LOG}"
