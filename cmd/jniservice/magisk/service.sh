#!/system/bin/sh
# jniservice auto-start via Magisk.
# Runs after boot, starts the gRPC server via app_process.

MODDIR="${0%/*}"
JNISERVICE_DIR="$MODDIR/jniservice"
LOG="/data/local/tmp/jniservice.log"

# Wait for boot to fully complete.
while [ "$(getprop sys.boot_completed)" != "1" ]; do
    sleep 1
done
sleep 5

# Read config from /data/local/tmp/jniservice.env if it exists.
PORT="${JNISERVICE_PORT:-50051}"
LISTEN="${JNISERVICE_LISTEN:-0.0.0.0}"
TOKEN="${JNISERVICE_TOKEN:-}"
if [ -f /data/local/tmp/jniservice.env ]; then
    . /data/local/tmp/jniservice.env
fi

# Kill any existing instance.
pkill -f "app_process.*JNIService" 2>/dev/null
sleep 1

# Start jniservice.
export JNISERVICE_PORT="$PORT"
export JNISERVICE_LISTEN="$LISTEN"
export JNISERVICE_TOKEN="$TOKEN"
export LD_LIBRARY_PATH="$JNISERVICE_DIR:$LD_LIBRARY_PATH"

app_process \
    -Djava.class.path="$JNISERVICE_DIR/jniservice.dex" \
    "$JNISERVICE_DIR" JNIService \
    >> "$LOG" 2>&1 &

echo "jniservice started (pid=$!, port=$PORT)" >> "$LOG"
