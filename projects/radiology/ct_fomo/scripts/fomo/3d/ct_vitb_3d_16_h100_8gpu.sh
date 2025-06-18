#!/bin/bash
# SLURM SUBMIT SCRIPT
#SBATCH --tasks-per-node=8
#SBATCH --job-name=ct_vitb
#SBATCH --gres=gpu:8
#SBATCH --partition=h100
#SBATCH --qos=eight_h100_qos
#SBATCH --nodelist=herakles
#SBATCH --cpus-per-task=16
#SBATCH --mem=0
#SBATCH --time=7-00:00:00
#SBATCH --output=ct_vitb_3d_%A_%a.out
#SBATCH --error=ct_vitb_3d_%A_%a.err

if [ -z "$1" ]; then
  echo "Please provide a url to a sqlite db containing CT-image metadata."
  exit 1
fi

DATABASE_URL=$1
PROJECT_ROOT=${2:-/projects/ct_SSL_encoder/aiforoncology/aiforoncology-internal/}

cd $PROJECT_ROOT/aifo/fomo
echo "Starting data producing"
bazelisk run :start_data_producer_vector -- \
  --name ct_vector_dataset --producer-type ct --chunk-size-mb 25 --max-memory-size-gb 700 \
  --num-workers 12 --database-url $DATABASE_URL \
  --slices-per-chunk 24 >$SCRATCH/producer_vitb16.out 2>$SCRATCH/producer_vitb16.err &

sleep 200 # wait for the data producer to start

date_now=$(date '+%Y%m%d-%H%M%S')
echo "Starting training"
bazelisk run :dinov2_train_ct_3d.venv
source .aifo+fomo+dinov2_train_ct_3d.venv/bin/activate

srun --export=ALL python fomo/ct/train/train3d.py \
  --config-file fomo/ct/configs/train/3d/ct_vitb_3d_16_no_pretrain.yaml \
  --output-dir $SCRATCH/dino_output/3d/ct_vitb_16_$date_now \
  --no-resume
