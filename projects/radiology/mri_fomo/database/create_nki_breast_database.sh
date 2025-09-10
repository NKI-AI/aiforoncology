#!/bin/bash
# SLURM SUBMIT SCRIPT - Create NKI Breast database
#SBATCH --ntasks=1
#SBATCH --job-name=nki_breast_database
#SBATCH --mem=32G
#SBATCH --qos=cpu_qos
#SBATCH --partition=cpu
#SBATCH --nodelist=gaia
#SBATCH --cpus-per-task=4
#SBATCH --time=2-00:00:00
#SBATCH --output=/home/v.v.veldhuizen/slurm_output/mri_fomo/database/nki_breast_database_%A.out
#SBATCH --error=/home/v.v.veldhuizen/slurm_output/mri_fomo/database/nki_breast_database_%A.err

# Script to create NKI Breast database containing all .nrrd files
# This script will crawl /data/groups/aiforoncology/derived/radiology/nki_breast/
# and create a database named nki_breast_300525_philips+200siemens_all_mods.sqlite

set -e # Exit on any error

# Change to the project root directory
cd /home/v.v.veldhuizen/aiforoncology-internal

echo "Creating NKI Breast database..."
echo "Source directory: /projects/nki-breast-mri/snapshot/aprep111/data/"
echo "Database name: aprep111_data.sqlite"
echo "Configuration: aifo/fomo/fomo/mri/configs/dataset/aprep111.yaml"
echo ""

# Run the database creation with bazelisk
bazelisk run //aifo/fomo:make_database_generic -- \
  --dataset-config /home/v.v.veldhuizen/aiforoncology-internal/aifo/fomo/fomo/mri/configs/dataset/aprep111_dataset_config.yaml \
  --database-path /projects/mri_fomo/database/aprep111_data.sqlite \
  --batch-size 100 \
  --num-workers 4 \
  --modality mri

echo ""
echo "Database creation completed!"
echo "Database saved as: /projects/mri_fomo/database/aprep111_data.sqlite"
