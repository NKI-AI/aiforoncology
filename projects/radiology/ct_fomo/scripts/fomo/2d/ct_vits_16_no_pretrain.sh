#!/bin/bash
# SLURM SUBMIT SCRIPT
#SBATCH --tasks-per-node=4
#SBATCH --job-name=ct_vits
#SBATCH --gres=gpu:4
#SBATCH --partition=a100
#SBATCH --qos=four_a100_qos
#SBATCH --nodelist=euctemon
#SBATCH --cpus-per-task=16
#SBATCH --time=7-00:00:00
#SBATCH --output=ct_vits.out
#SBATCH --error=ct_vits.err

if [ -z "$1" ]; then
  echo "Please provide a url to a sqlite db containing CT-image metadata."
  exit 1
fi

DATABASE_URL=$1
PROJECT_ROOT=${2:-/projects/ct_SSL_encoder/aiforoncology/aiforoncology-internal}

cd $PROJECT_ROOT/aifo/fomo
echo "Starting data producing"
bazelisk run :start_data_producer_vector -- \
  --name ct_vector_dataset --producer-type ct --chunk-size-mb 3 --max-memory-size-gb 256 \
  --num-workers 16 --database-url $DATABASE_URL \
  >$SCRATCH/producer_vits16.out 2>$SCRATCH/producer_vits16.err &

sleep 120 # wait for the data producer to start

date_now=$(date '+%Y%m%d-%H%M%S')
echo "Starting training"
bazelisk run :dinov2_train_ct_2d.venv
source .aifo+fomo+dinov2_train_ct_2d.venv/bin/activate
srun --export=ALL python fomo/ct/train/train_2d.py \
  --config-file fomo/ct/configs/train/2d/ct_vits_16_no_pretrain.yaml \
  --output-dir $SCRATCH/dino_output/2d/ct_vits_16_$date_now \
  --no-resume
