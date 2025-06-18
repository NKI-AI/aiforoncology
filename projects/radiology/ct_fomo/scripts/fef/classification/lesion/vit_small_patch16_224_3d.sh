#!/bin/bash
#SBATCH --tasks-per-node=1
#SBATCH --job-name=lesion_vits_3d
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
  echo "Please provide a path to a timm-converted 3d ViT checkpoint."
  exit 1
fi

EVAL_NAME=$1
CKPT_PATH=$2
PROJECT_ROOT=${3:-/projects/ct_SSL_encoder/aiforoncology/aiforoncology-internal}

export MODEL_NAME=vit_small_patch16_224_3d
export CKPT_PATH=$CKPT_PATH
export EMBEDDINGS_ROOT=$SCRATCH/$EVAL_NAME/lesion/embeddings
export OUTPUT_ROOT=$SCRATCH/$EVAL_NAME/lesion/logs
export IN_FEATURES=384
# TODO: This should be configurable in the future when models are trained with different z-dimensional input
export CHUNK_SIZE=16
export DATA_ROOT=$SCRATCH/lesion
export RESIZE_DIM="[$CHUNK_SIZE, 224, 224]"
export RESIZE_MODE="trilinear"

cd $PROJECT_ROOT/aifo/fef
srun --export=ALL bazelisk run :eva -- predict_fit --config aifo/fef/fef/configs/ct/lesion.yaml
