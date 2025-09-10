from research.physician_gtv_fingerprint.data.components.ct_dataset import CTDataset

ds = CTDataset(
    data_dir="/projects/physician_gtv_fingerprint_data/m42_nii/phys_fp_crop/", splits_fn="all/0.json", split="combined"
)

phys_counts = {_: 0 for _ in range(100)}
for i in range(len(ds)):
    phys_counts[ds.manifest[i]["physician"]] += 1


# filter the physicians with zero counts
phys_counts = {k: v for k, v in phys_counts.items() if v > 0}

print(phys_counts)

real_phys_counts = {}
for k in phys_counts.keys():
    patients = [ds.manifest[i] for i in range(len(ds)) if ds.manifest[i]["physician"] == k]
    if k in ds.physicians:
        real_phys_counts[k] = len(patients)
    else:
        if 100 not in real_phys_counts:
            real_phys_counts[100] = 0
        real_phys_counts[100] += len(patients)
    # print(f"Physician {k} has {len(patients)} patients")
    # print(f"number of unique patients: {len(set([p['AnonID'] for p in patients]))}")

print(real_phys_counts)
exit()

pt_scan_counts = {}
pt_physicians = {}
for i in range(len(ds)):
    anon_id = ds.manifest[i]["AnonID"]
    if anon_id not in pt_scan_counts:
        pt_scan_counts[anon_id] = 0
    if anon_id not in pt_physicians:
        pt_physicians[anon_id] = []

    pt_scan_counts[anon_id] += 1
    pt_physicians[anon_id].append(ds.manifest[i]["physician"])

# get all patients with more than one scan
multi_scan_patients = {k: v for k, v in pt_scan_counts.items() if v > 1}
print(f"Number of patients with more than one scan: {len(multi_scan_patients)}")

# for each of those patients, get the list of physicians
for k in multi_scan_patients.keys():
    print(f"Patient {k} has {pt_scan_counts[k]} scans")
    print(f"Physicians: {set(pt_physicians[k])}")
    print("")
