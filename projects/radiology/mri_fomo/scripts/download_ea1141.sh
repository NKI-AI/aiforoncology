#!/bin/bash
# SLURM SUBMIT SCRIPT
#SBATCH --tasks-per-node=1
#SBATCH --job-name=download_ea1141
#SBATCH --partition=cpu
#SBATCH --nodelist=gaia
#SBATCH --cpus-per-task=1
#SBATCH --qos=cpu_qos
#SBATCH --time=7-00:00:00
#SBATCH --output=/home/v.v.veldhuizen/slurm_output/download_ea1141_%A.out
#SBATCH --error=/home/v.v.veldhuizen/slurm_output/download_ea1141_%A.err

# Set up environment
NBIA_HOME="/home/v.v.veldhuizen/nbia-install"
NBIA_BASE_URL="https://services.cancerimagingarchive.net/services/v4"
COLLECTION="EA1141"
MODALITY="MR"

# Create output directory
OUTPUT_DIR="/data/groups/public/archive/radiology/ea1141"
mkdir -p $OUTPUT_DIR

# Logging function
log() {
  echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1"
}

# Cleanup function
cleanup() {
  local dir="$1"
  log "Cleaning up temporary files in $dir"
  rm -f "$dir/images.zip.tmp" "$dir/images.zip"
}

# Check if patient is already downloaded
is_patient_downloaded() {
  local dir="$1"
  # Check if directory exists and has files (excluding temp files)
  if [ -d "$dir" ] && [ "$(ls -A "$dir" | grep -v '\.tmp$' | wc -l)" -gt 0 ]; then
    return 0
  fi
  return 1
}

# Debug function to check file contents
debug_file() {
  local file="$1"
  log "Debugging file: $file"
  log "File size: $(ls -lh "$file" | awk '{print $5}')"
  log "First 100 bytes: $(head -c 100 "$file" | xxd -p)"
  log "File type: $(file "$file")"
}

# Download images for a series
download_series() {
  local series_uid="$1"
  local output_dir="$2"
  local max_retries=3
  local retry_count=0

  while [ $retry_count -lt $max_retries ]; do
    log "Downloading images for series $series_uid (attempt $((retry_count + 1))/$max_retries)..."
    local temp_file="$output_dir/images.zip.tmp"
    local final_file="$output_dir/images.zip"

    # Use the getImage endpoint
    local api_url="$NBIA_BASE_URL/TCIA/query/getImage?SeriesInstanceUID=$series_uid"
    log "API URL: $api_url"

    curl --max-time 300 -s "$api_url" --output "$temp_file"
    local curl_status=$?

    if [ $curl_status -ne 0 ]; then
      log "ERROR: Curl command failed with status $curl_status"
      cleanup "$output_dir"
      ((retry_count++))
      continue
    fi

    # Check if file was downloaded
    if [ ! -s "$temp_file" ]; then
      log "ERROR: Downloaded file is empty"
      cleanup "$output_dir"
      ((retry_count++))
      continue
    fi

    # Debug the downloaded file
    debug_file "$temp_file"

    # Check if it's a valid ZIP file
    if ! unzip -t "$temp_file" >/dev/null 2>&1; then
      log "ERROR: Downloaded file is not a valid ZIP file"
      log "File contents:"
      head -n 5 "$temp_file"
      cleanup "$output_dir"
      ((retry_count++))
      continue
    fi

    # Move temp file to final location
    mv "$temp_file" "$final_file"

    log "Unzipping images for series $series_uid..."
    if ! unzip -q "$final_file" -d "$output_dir"; then
      log "ERROR: Failed to unzip images"
      cleanup "$output_dir"
      ((retry_count++))
      continue
    fi

    rm "$final_file"
    log "Successfully downloaded and extracted images for series $series_uid"
    return 0
  done

  log "ERROR: Failed to download series $series_uid after $max_retries attempts"
  cleanup "$output_dir"
  return 1
}

# Download patient data
download_patient() {
  local patient_id="$1"
  local output_dir="$2"

  # Check if patient is already downloaded
  if is_patient_downloaded "$output_dir"; then
    log "Patient $patient_id already downloaded, skipping"
    return 0
  fi

  log "Creating directory for patient $patient_id: $output_dir"
  mkdir -p "$output_dir"

  # Get studies for the patient
  log "Getting studies for patient $patient_id..."
  local studies_url="$NBIA_BASE_URL/TCIA/query/getPatientStudy?Collection=$COLLECTION&PatientID=$patient_id&format=json"
  local studies_response=$(curl -s "$studies_url")

  if [ -z "$studies_response" ]; then
    log "ERROR: No studies found for patient $patient_id"
    return 1
  fi

  log "Studies response: $studies_response"

  # Extract StudyInstanceUIDs
  local study_uids=$(echo "$studies_response" | grep -o '"StudyInstanceUID":"[^"]*"' | cut -d'"' -f4)

  if [ -z "$study_uids" ]; then
    log "ERROR: No StudyInstanceUIDs found in response"
    return 1
  fi

  # Process each study
  for study_uid in $study_uids; do
    log "Processing study $study_uid..."

    # Get series for the study
    local series_url="$NBIA_BASE_URL/TCIA/query/getSeries?Collection=$COLLECTION&StudyInstanceUID=$study_uid&Modality=$MODALITY&format=json"
    local series_response=$(curl -s "$series_url")

    if [ -z "$series_response" ]; then
      log "No MR series found for study $study_uid, skipping"
      continue
    fi

    log "Series response: $series_response"

    # Extract SeriesInstanceUIDs
    local series_uids=$(echo "$series_response" | grep -o '"SeriesInstanceUID":"[^"]*"' | cut -d'"' -f4)

    if [ -z "$series_uids" ]; then
      log "ERROR: No SeriesInstanceUIDs found in response"
      continue
    fi

    # Process each series
    for series_uid in $series_uids; do
      log "Processing series $series_uid..."
      local series_dir="$output_dir/$study_uid/$series_uid"
      mkdir -p "$series_dir"

      if ! download_series "$series_uid" "$series_dir"; then
        log "ERROR: Failed to download series $series_uid"
        continue
      fi
    done
  done

  return 0
}

# Start the download process
log "Starting download of $COLLECTION collection (MR studies only)"

# Get list of patients
log "Getting patient list..."
patients_url="$NBIA_BASE_URL/TCIA/query/getPatient?Collection=$COLLECTION&format=json"
patients_response=$(curl -s "$patients_url")

if [ -z "$patients_response" ]; then
  log "ERROR: Failed to get patient list"
  exit 1
fi

log "Patients response: $patients_response"

# Extract all patient IDs
patient_ids=$(echo "$patients_response" | grep -o '"PatientID":"[^"]*"' | cut -d'"' -f4)
total_patients=$(echo "$patient_ids" | wc -l)

if [ -z "$patient_ids" ]; then
  log "ERROR: No patient IDs found in response"
  log "Full response: $patients_response"
  exit 1
fi

log "Found $total_patients patients in collection $COLLECTION"
log "Patient IDs: $patient_ids"

# Download all patients
current_patient=1
for patient_id in $patient_ids; do
  log "Processing patient $current_patient of $total_patients: $patient_id"
  if ! download_patient "$patient_id" "$OUTPUT_DIR/$patient_id"; then
    log "WARNING: Failed to download patient $patient_id, continuing with next patient"
  fi
  ((current_patient++))
done

log "Download completed"
