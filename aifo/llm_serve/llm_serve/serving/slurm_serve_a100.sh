#!/bin/bash
#SBATCH --job-name=llm_serve_a100
#SBATCH --partition=a100
#SBATCH --qos=eight_a100_qos
#SBATCH --gres=gpu:1
#SBATCH --cpus-per-task=16
#SBATCH --mem=128G
#SBATCH --nodelist=eudoxus
#SBATCH --time=3-00:00:00
#SBATCH --output=llm_serve.log
#SBATCH --error=llm_serve.err

# Model configuration
MODEL_DIR="/projects/public_llms/qwen32bdeepseek"
PORT=8030
WEB_PORT=8050

# Start vLLM server
bazelisk run //aifo/llm_serve:api_server -- \
  --model $MODEL_DIR \
  --served-model-name qwen32bdeepseek \
  --tensor-parallel-size 1 \
  --gpu-memory-utilization 0.99 \
  --dtype half \
  --max-model-len 32768 \
  --host 0.0.0.0 \
  --port $PORT &

# Health check
echo "Waiting for server to start..."
sleep 150
curl -sS http://localhost:$PORT/v1/models >/dev/null
if [ $? -eq 0 ]; then
  echo "Server started successfully on port $PORT"
else
  echo "Failed to start server"
  exit 1
fi

# Start web server
bazelisk run //aifo/llm_serve:server -- \
  --port $WEB_PORT \
  --host 0.0.0.0 &

echo "Web server is running"
wait
