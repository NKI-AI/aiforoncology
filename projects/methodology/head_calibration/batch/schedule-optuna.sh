#!/bin/bash
#SBATCH --partition a6000
#SBATCH --qos a6000_qos
#SBATCH --gpus 1
#SBATCH --time 6:0:0

hostname
echo $SLURM_JOB_ID
echo ${USER}
source /home/${USER}/.bashrc
pwd
source ~/miniconda3/etc/profile.d/conda.sh
conda activate headcal

python src/train.py -m hydra/launcher=submitit_local hparams_search=xxx data.num_workers=8 \
  +trainer.enable_progress_bar=False experiment=xxx hydra.launcher.gpus_per_node=1
