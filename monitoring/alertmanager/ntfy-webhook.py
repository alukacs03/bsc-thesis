#!/usr/bin/env python3
import json
import os
import logging
from http.server import HTTPServer, BaseHTTPRequestHandler
from urllib.request import Request, urlopen
from urllib.error import URLError

NTFY_URL = os.environ.get('NTFY_URL', 'https://notify.numnet.eu')
NTFY_TOPIC = os.environ.get('NTFY_TOPIC', 'gluon-alerts')
NTFY_CRITICAL_TOPIC = os.environ.get('NTFY_CRITICAL_TOPIC', 'gluon-critical')
LISTEN_PORT = int(os.environ.get('LISTEN_PORT', '9095'))
NTFY_TOKEN = os.environ.get('NTFY_TOKEN', '')

logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

PRIORITY_MAP = {
    'critical': '5',
    'warning': '4',
    'info': '3',
}

STATUS_EMOJI = {
    'firing': '🔥',
    'resolved': '✅',
}

SEVERITY_EMOJI = {
    'critical': '🚨',
    'warning': '⚠️',
    'info': 'ℹ️',
}


def format_alert_message(alert: dict, status: str) -> str:
    labels = alert.get('labels', {})
    annotations = alert.get('annotations', {})

    alertname = labels.get('alertname', 'Unknown')
    severity = labels.get('severity', 'info')
    summary = annotations.get('summary', alertname)
    description = annotations.get('description', '')

    status_emoji = STATUS_EMOJI.get(status, '❓')
    severity_emoji = SEVERITY_EMOJI.get(severity, '')

    lines = [
        f"{status_emoji} {severity_emoji} {summary}",
    ]

    if description:
        lines.append(f"\n{description}")

    skip_labels = {'alertname', 'severity', 'job', 'instance'}
    extra_labels = {k: v for k, v in labels.items() if k not in skip_labels}
    if extra_labels:
        label_str = ', '.join(f"{k}={v}" for k, v in extra_labels.items())
        lines.append(f"\nLabels: {label_str}")

    return '\n'.join(lines)


def send_to_ntfy(title: str, message: str, priority: str, topic: str, tags: list):
    url = f"{NTFY_URL}/{topic}"

    headers = {
        'Title': title,
        'Priority': priority,
        'Tags': ','.join(tags),
    }

    if NTFY_TOKEN:
        headers['Authorization'] = f'Bearer {NTFY_TOKEN}'

    try:
        req = Request(url, data=message.encode('utf-8'), headers=headers, method='POST')
        with urlopen(req, timeout=10) as response:
            logger.info(f"Sent to ntfy: {response.status} - {title}")
            return True
    except URLError as e:
        logger.error(f"Failed to send to ntfy: {e}")
        return False


class WebhookHandler(BaseHTTPRequestHandler):
    def do_POST(self):
        content_length = int(self.headers.get('Content-Length', 0))
        body = self.rfile.read(content_length)

        try:
            payload = json.loads(body.decode('utf-8'))
        except json.JSONDecodeError as e:
            logger.error(f"Invalid JSON: {e}")
            self.send_response(400)
            self.end_headers()
            return

        logger.info(f"Received webhook: status={payload.get('status')}, alerts={len(payload.get('alerts', []))}")

        status = payload.get('status', 'unknown')
        alerts = payload.get('alerts', [])

        for alert in alerts:
            labels = alert.get('labels', {})
            severity = labels.get('severity', 'info')
            alertname = labels.get('alertname', 'Alert')

            topic = NTFY_CRITICAL_TOPIC if severity == 'critical' else NTFY_TOPIC

            message = format_alert_message(alert, status)

            status_text = 'FIRING' if status == 'firing' else 'RESOLVED'
            title = f"[{status_text}] {alertname}"

            priority = PRIORITY_MAP.get(severity, '3')
            if status == 'resolved':
                priority = '3'

            tags = [severity, status]
            if status == 'firing' and severity == 'critical':
                tags.append('rotating_light')
            elif status == 'resolved':
                tags.append('white_check_mark')

            send_to_ntfy(title, message, priority, topic, tags)

        self.send_response(200)
        self.send_header('Content-Type', 'application/json')
        self.end_headers()
        self.wfile.write(b'{"status": "ok"}')

    def do_GET(self):
        if self.path == '/health':
            self.send_response(200)
            self.send_header('Content-Type', 'application/json')
            self.end_headers()
            self.wfile.write(b'{"status": "healthy"}')
        else:
            self.send_response(404)
            self.end_headers()

    def log_message(self, format, *args):
        pass


def main():
    server = HTTPServer(('0.0.0.0', LISTEN_PORT), WebhookHandler)
    logger.info(f"Starting ntfy webhook relay on port {LISTEN_PORT}")
    logger.info(f"ntfy URL: {NTFY_URL}")
    logger.info(f"Topics: {NTFY_TOPIC} (default), {NTFY_CRITICAL_TOPIC} (critical)")
    try:
        server.serve_forever()
    except KeyboardInterrupt:
        logger.info("Shutting down...")
        server.shutdown()


if __name__ == '__main__':
    main()
