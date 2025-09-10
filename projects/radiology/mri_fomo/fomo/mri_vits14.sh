#!/bin/bash
# SLURM SUBMIT SCRIPT
#SBATCH --tasks-per-node=4
#SBATCH --job-name=mri_vits
#SBATCH --gres=gpu:4
#SBATCH --partition=a100
#SBATCH --qos=eight_a100_qos
#SBATCH --nodelist=eudoxus
#SBATCH --cpus-per-task=16
#SBATCH --time=7-00:00:00
#SBATCH --output=mri_vits_%A_%a.out
#SBATCH --error=mri_vits_%A_%a.err

if [ -z "$1" ]; then
  echo "Please provide a url to a sqlite db containing MRI-image metadata."
  exit 1
fi

DATABASE_URL=$1
PROJECT_ROOT=${2:-/projects/mri_fomo/aiforoncology/aiforoncology-internal}
current_user=$(whoami)

cd $PROJECT_ROOT/aifo/fomo
echo "Starting data producing"
bazelisk run :start_data_producer_vector -- \
  --name mri_vector --producer-type mri --chunk-size-mb 5 --max-memory-size-gb 250 \
  --num-workers 16 --database-url $DATABASE_URL \
  >/processing/$current_user/producer_vits.out 2>/processing/$current_user/producer_vits.err &

sleep 120 # wait for the data producer to start

date_now=$(date '+%Y%m%d-%H%M%S')
echo "Starting training"
srun --export=ALL bazelisk run :dinov2_train_mri -- \
  --config-file fomo/mri/configs/train/mri_vits14.yaml \
  --output-dir /processing/$current_user/dino_output/$date_now
