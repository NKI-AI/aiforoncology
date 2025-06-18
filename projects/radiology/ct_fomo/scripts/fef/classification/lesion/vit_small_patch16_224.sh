#!/bin/bash
#SBATCH --tasks-per-node=1
#SBATCH --job-name=lesion_vits
#SBATCH --gres=gpu:1
#SBATCH --partition=a6000
#SBATCH --qos=a6000_qos
#SBATCH --nodelist=galileo
#SBATCH --cpus-per-task=16
#SBATCH --time=7-00:00:00
#SBATCH --output=lesion_%A_%a.out
#SBATCH --error=lesion_%A_%a.err

if [ -z "$1" ]; then
  echo "Please give a name to this evaluation"
  exit 1
fi

if [ -z "$2" ]; then
  echo "Please provide a path to a timm-converted 2d ViT checkpoint."
  exit 1
fi

EVAL_NAME=$1
CKPT_PATH=$2
PROJECT_ROOT=${3:-/projects/ct_SSL_encoder/aiforoncology/aiforoncology-internal}

export MODEL_NAME=vit_small_patch16_224
export CKPT_PATH=$CKPT_PATH
export EMBEDDINGS_ROOT=$SCRATCH/$EVAL_NAME/lesion/embeddings
export OUTPUT_ROOT=$SCRATCH/$EVAL_NAME/lesion/logs
export IN_FEATURES=384
export CHUNK_SIZE=1
export DATA_ROOT=$SCRATCH/lesion

cd $PROJECT_ROOT/aifo/fef
srun --export=ALL bazelisk run :eva -- predict_fit --config aifo/fef/fef/configs/ct/lesion.yaml
