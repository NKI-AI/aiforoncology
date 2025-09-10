#!/bin/bash
# SLURM SUBMIT SCRIPT
#SBATCH --tasks-per-node=4
#SBATCH --job-name=mri_fomo_eval
#SBATCH --gres=gpu:1
#SBATCH --partition=a100
#SBATCH --qos=eight_a100_qos
#SBATCH --cpus-per-task=4
#SBATCH --time=2-00:00:00
#SBATCH --output=/home/v.v.veldhuizen/aiforoncology-internal/output/mri_fomo_eval_%A_%a.out
#SBATCH --error=/home/v.v.veldhuizen/aiforoncology-internal/output/mri_fomo_eval_%A_%a.err

source /home/v.v.veldhuizen/miniforge3/etc/profile.d/conda.sh
conda activate mf

cd /home/v.v.veldhuizen/aiforoncology-internal/
export PYTHONPATH=$PYTHONPATH:/home/v.v.veldhuizen/aiforoncology-internal/third_party/eva/src:/home/v.v.veldhuizen/aiforoncology-internal/src/python

export IN_FEATURES=384
export MODEL_NAME="mri_vits14_no_pretrain_v3"
export BATCH_SIZE=1
export LABEL_FILE="/data/groups/public/archive/MAMA-MIA/manifests/manifest_duke.csv"
export LABEL_NAME='tumor_subtype'
export DATASET_NAME="duke"
export MAX_STEPS=500000
eva fit --config /home/v.v.veldhuizen/aiforoncology-internal/src/python/research/fomo/mri/eval/configs/classification/tumor_type/subtype_mil_single.yaml
