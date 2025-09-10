#!/bin/bash
#SBATCH --tasks-per-node=1
#SBATCH --job-name=kits23
#SBATCH --gres=gpu:1
#SBATCH --partition=a6000
#SBATCH --qos=a6000_qos
#SBATCH --cpus-per-task=16
#SBATCH --time=1-00:00:00
#SBATCH --output=kits23_%A_%a.out
#SBATCH --error=kits23_%A_%a.err

PROJECT_ROOT=${PROJECT_ROOT:-/projects/ct_SSL_encoder/aiforoncology/aiforoncology-internal}
export MODEL_NAME=$MODEL_NAME
export CKPT_PATH=$CKPT_PATH
export EMBEDDINGS_ROOT=$SCRATCH/downstream_task_results/$EVAL_NAME/kits23/embeddings
export OUTPUT_ROOT=$SCRATCH/downstream_task_results/$EVAL_NAME/kits23/logs
export IN_FEATURES=$IN_FEATURES
export DATA_ROOT=$SCRATCH/kits23

echo 'Running eval with settings:'
echo "MODEL_NAME=$MODEL_NAME"
echo "CKPT_PATH=$CKPT_PATH"
echo "EMBEDDINGS_ROOT=$EMBEDDINGS_ROOT"
echo "OUTPUT_ROOT=$OUTPUT_ROOT"
echo "IN_FEATURES=$IN_FEATURES"
echo "DATA_ROOT=$DATA_ROOT"

cd $PROJECT_ROOT/aifo/fef
srun --export=ALL bazelisk run :eva -- predict_fit --config aifo/fef/fef/configs/ct/2d/kits23.yaml
echo "Removing embeddings located at '$EMBEDDINGSROOT'."
rm -rf $EMBEDDINGS_ROOT
echo "Removal complete."
