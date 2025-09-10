import os
from typing import Dict, List, Optional, Tuple

import yaml
from bokeh.io import curdoc
from bokeh.layouts import Spacer, column, layout, row
from bokeh.models import Button, Div, Select
from research.dicom_annotator.database import session
from research.dicom_annotator.labels_database import SeriesLabel, init_labels_db
from research.dicom_annotator.models import Instances, MriSpecifics, Patients, Series, Studies
from research.dicom_annotator.utils import MriParameters, MriSequenceClassifier

# Define the YAML configuration at the module level
LABEL_CONFIG_YAML = """
labels:
  - T1w multivolume
  - T1w (no contrast)
  - T1w unknown
  - T1w (with contrast):
      phases: 15
  - Subtraction T1w:
      phases: 15
  - T2w
  - DWI
  - ADC
  - Localizer
  - Unknown/Other
"""

LABELS_DATABASE_PATH = os.environ.get("LABELS_DATABASE_PATH", None)
if LABELS_DATABASE_PATH is None:
    raise ValueError("Please set the LABELS_DATABASE_PATH environment variable.")


class LabelManager:
    def __init__(self, config_yaml: str):
        self.config = yaml.safe_load(config_yaml)
        self.label_options, self.phase_labels = self._generate_label_options()
        self.labels_session = init_labels_db("sqlite:///{0}".format(LABELS_DATABASE_PATH))
        print("Label database path:", LABELS_DATABASE_PATH)

    def _generate_label_options(self) -> Tuple[List[str], Dict[str, List[str]]]:
        label_options = []
        phase_labels = {}
        for label in self.config["labels"]:
            if isinstance(label, str):
                label_options.append(label)
            elif isinstance(label, dict):
                main_label = list(label.keys())[0]
                label_options.append(main_label)
                if "phases" in label[main_label]:
                    phase_labels[main_label] = ["N/A"] + [str(i) for i in range(1, label[main_label]["phases"] + 1)]
        return label_options, phase_labels

    def create_label_select(self, series: Series) -> Tuple[Select, Select]:
        existing_label = (
            self.labels_session.query(SeriesLabel).filter_by(series_instance_uid=series.series_instance_uid).first()
        )

        if existing_label:
            if "//" in existing_label.series_type:
                initial_label, phase_part = existing_label.series_type.split("//")
                initial_phase = phase_part.replace("Phase ", "").strip()
            else:
                initial_label = existing_label.series_type
                initial_phase = "N/A"
            label_select = Select(title="Label", options=self.label_options, value=initial_label, width=150)
            label_select.styles = {
                "background": "#e3f2fd",
                "color": "#1565c0",
                "font-weight": "bold",
                "border": "1px solid #90caf9",
            }
        else:
            label_select = Select(title="Label", options=self.label_options, value="", width=150)
            initial_label = self.label_options[0]
            initial_phase = "N/A"

        phase_select = Select(
            title="Phase",
            options=self.phase_labels.get(initial_label, []),
            value=initial_phase,
            visible=initial_label in self.phase_labels,
            width=60,
        )

        def update_phase_visibility(attr, old, new):
            phase_select.visible = new in self.phase_labels
            if new in self.phase_labels:
                phase_select.options = self.phase_labels[new]
                if phase_select.value not in phase_select.options:
                    phase_select.value = "N/A"  # Default to "N/A"

        label_select.on_change("value", update_phase_visibility)

        return label_select, phase_select

    def save_label(self, series: Series, label: str, phase: Optional[str] = None) -> None:
        series_type = f"{label}//Phase {phase}" if phase else label
        existing_label = (
            self.labels_session.query(SeriesLabel).filter_by(series_instance_uid=series.series_instance_uid).first()
        )
        if existing_label:
            existing_label.series_description = series.series_description or "No Description"
            existing_label.series_type = series_type
        else:
            new_label = SeriesLabel(
                series_instance_uid=series.series_instance_uid,
                series_description=series.series_description or "No Description",
                study_instance_uid=series.studies.study_instance_uid,
                study_description=series.studies.study_description or "No Description",
                series_type=series_type,
            )
            self.labels_session.add(new_label)


class DicomSeriesLabeler:
    def __init__(self):
        self.label_manager = LabelManager(LABEL_CONFIG_YAML)
        self.sequence_classifier = MriSequenceClassifier()
        self.current_series_entries = []
        self.current_patient_index = 0
        self.current_study_index = 0
        self.prev_time = None
        self.setup_ui()

    def setup_ui(self):
        self.create_widgets()
        self.setup_callbacks()
        self.create_layout()
        self.initialize_data()

    def create_widgets(self):
        patients = self.get_patients()
        patient_options = [(str(p.id), p.patient_name) for p in patients]

        self.patient_select = Select(title="Select Patient", options=patient_options)
        self.study_select = Select(title="Select Study", options=[])
        self.series_list_column = column()
        self.series_info_div = Div(
            text="<h3>Series Information</h3><p>Select a series to view more details here.</p>", width=400
        )
        self.prev_study_button = Button(label="<<", button_type="success", width=50, height=30)
        self.next_study_button = Button(label=">>", button_type="success", width=50, height=30)
        self.save_button = Button(label="Save", button_type="primary", width=100, height=30)

    def setup_callbacks(self):
        self.patient_select.on_change("value", lambda attr, old, new: self.load_patient_studies(int(new)))
        self.study_select.on_change("value", lambda attr, old, new: self.load_study_series(int(new)))
        self.prev_study_button.on_click(self.go_to_previous_study)
        self.next_study_button.on_click(self.go_to_next_study)
        self.save_button.on_click(self.save_current_selection)

    def create_layout(self):
        self.layout = layout(
            [
                [Div(text="<h2>DICOM Series Labeling</h2>")],
                row(
                    column(
                        Div(text="<h3>Patient and Study Selection</h3>"),
                        self.patient_select,
                        Spacer(height=10),
                        self.study_select,
                        Spacer(height=20),
                        self.save_button,
                        row(self.prev_study_button, self.next_study_button),
                    ),
                    Spacer(width=50),
                    column(Div(text="<h3>Series List</h3>"), self.series_list_column, Spacer(height=20)),
                    Spacer(width=50),
                    column(self.series_info_div),
                ),
            ]
        )

    def initialize_data(self):
        patients = self.get_patients()
        if patients:
            self.patient_select.value = str(patients[0].id)
            self.load_patient_studies(patients[0].id)

    def get_patients(self):
        return session.query(Patients).all()

    def load_patient_studies(self, patient_id: int):
        studies = session.query(Studies).filter_by(patient_id=patient_id).all()
        study_options = [(str(study.id), study.study_description or "No Description") for study in studies]
        self.study_select.options = study_options
        if study_options:
            self.study_select.value = study_options[0][0]
            self.load_study_series(int(study_options[0][0]))

    def calculate_time_difference(self, series: Series) -> str:
        if self.prev_time is None:
            time_diff = "0s"
        else:
            diff_seconds = int((series.series_date_time - self.prev_time).total_seconds())
            time_diff = f"+{diff_seconds}s"
        self.prev_time = series.series_date_time
        return time_diff

    def create_series_row(self, series: Series) -> Tuple[int, Select, Select, row]:
        time_diff = self.calculate_time_difference(series)
        label_select, phase_select = self.label_manager.create_label_select(series)

        # Get instance count
        instance_count = session.query(Instances).filter(Instances.series_id == series.id).count()

        # Format series description
        desc = series.series_description or "No Description"

        metadata = f"{time_diff} (#{instance_count})"

        # Create two-line label
        button_label = f"{desc} {metadata}"

        series_button = Button(
            label=button_label,
            width=350,
            height=50,
            button_type="light",
        )

        # Style the button to look more like text
        series_button.styles = {
            "background": "white",  # White background
            "border": "1px solid #e0e0e0",  # Light grey border
            "box-shadow": "none",  # Remove default button shadow
            "padding": "5px 10px",  # Add some padding
            "text-align": "left",  # Left align text
            "white-space": "pre-line",  # Preserve line breaks
            "font-family": "sans-serif",  # Use sans-serif font
            "line-height": "1.2",  # Adjust line height
            "font-size": "14px",  # Base font size for description
            "background-image": "none",  # Remove any button gradient
            "text-transform": "none",  # Prevent uppercase transformation
            "color": "#000000",  # Black text for description
        }

        series_button.on_click(lambda: self.display_series_info(series.id))

        return (series.id, label_select, phase_select, row(series_button, row(label_select, phase_select, width=200)))

    def create_series_row2(self, series: Series) -> Tuple[int, Select, Select, row]:
        time_diff = self.calculate_time_difference(series)
        label_select, phase_select = self.label_manager.create_label_select(series)

        # Get instance count
        instance_count = session.query(Instances).filter(Instances.series_id == series.id).count()

        # Format series description
        desc = series.series_description or "No Description"

        # Format the button label with instance count
        button_label = f"{desc}\n{time_diff} ({instance_count} dcm)"

        series_button = Button(
            label=button_label,
            width=250,
            height=50,
            button_type="light",
            css_classes=["series-button"],
        )

        # Add custom styles to make the button more readable
        series_button.styles = {
            "white-space": "pre-wrap",  # Preserve line breaks
            "text-align": "left",  # Left align text
            "padding": "5px 10px",  # Add some padding
            "font-family": "monospace",  # Use monospace font for better alignment
            "line-height": "1.2",  # Adjust line height
        }

        series_button.on_click(lambda: self.display_series_info(series.id))

        return (series.id, label_select, phase_select, row(series_button, row(label_select, phase_select, width=200)))

    def load_study_series(self, study_id: int):
        series_list = session.query(Series).filter_by(study_id=study_id).order_by(Series.series_date_time).all()
        series_list = self.filter_series(series_list)
        self.prev_time = None  # Reset time difference calculation
        self.current_series_entries = [self.create_series_row(series) for series in series_list]
        self.series_list_column.children = [entry[-1] for entry in self.current_series_entries]

    # [Previous code remains the same up until the display_series_info method]

    def display_series_info(self, series_id: int):
        series = session.query(Series).filter(Series.id == series_id).first()
        if not series:
            self.series_info_div.text = "<h3>Series Information</h3><p>Series not found.</p>"
            return

        mri_instances = session.query(MriSpecifics).join(Series).filter(Series.id == series.id).all()
        instances = session.query(Instances).filter(Instances.series_id == series.id).all()

        if not mri_instances:
            self.series_info_div.text = "<p>No MRI instances found for this series.</p>"
            return

        # Get folder path from first instance
        folder = "/".join(instances[0].dicom_file_path.split("/")[:-1]) if instances else "N/A"

        # Extract MRI parameters from first instance
        first_mri = mri_instances[0]
        params = MriParameters(
            echo_time=float(first_mri.echo_time) if first_mri.echo_time else 0,
            repetition_time=float(first_mri.repetition_time) if first_mri.repetition_time else 0,
            inversion_time=float(first_mri.inversion_time) if first_mri.inversion_time else 0,
            flip_angle=float(first_mri.flip_angle) if first_mri.flip_angle else -1000,
            pixel_bandwidth=float(first_mri.pixel_bandwidth) if first_mri.pixel_bandwidth else None,
            image_type=first_mri.image_type,
        )

        guessed_sequence = self.sequence_classifier.guess_sequence_type(params)

        # Calculate slice location range
        slice_locations = [mri.slice_location for mri in instances if mri.slice_location is not None]
        if slice_locations:
            try:
                min_loc = min(map(float, slice_locations))
                max_loc = max(map(float, slice_locations))
                slice_range = abs(max_loc - min_loc)
            except ValueError:
                slice_range = "N/A"
        else:
            slice_range = "N/A"

        info_html = f"""
        <h3>Series Information</h3>
        <p><b>Description:</b> {series.series_description or "No Description"}</p>
        <p><b>Modality:</b> {series.modality or "N/A"}</p>
        <p><b>Protocol Name:</b> {series.protocol_name or "N/A"}</p>
        <p><b>Body Part:</b> {series.body_part_examined or "N/A"}</p>
        <p><b>Contrast Agent:</b> {series.contrast_bolus_agent or "N/A"}</p>
        <p><b>Manufacturer:</b> {series.manufacturer or "N/A"}</p>
        <p><b>Model:</b> {series.manufacturer_model_name or "N/A"}</p>
        <p><b>Date/Time:</b> {series.series_date_time or "N/A"}</p>
        <p><b>Slice Thickness:</b> {series.slice_thickness or "N/A"}</p>
        <p><b>Total Instances:</b> {len(instances)}</p>
        <p><b>MRI Instances:</b> {len(mri_instances)}</p>
        <p><b>Folder:</b> {folder}</p>
        <h3>Selected MRI Parameters</h3>
        <p><b>Echo Time:</b> {params.echo_time}</p>
        <p><b>Repetition Time:</b> {params.repetition_time}</p>
        <p><b>Inversion Time:</b> {params.inversion_time}</p>
        <p><b>Flip Angle:</b> {params.flip_angle}</p>
        <p><b>Total scan size:</b> {slice_range} mm</p>
        <p style="color: red;"><b>Guessed MRI Sequence Type:</b> {guessed_sequence}</p>
        """
        self.series_info_div.text = info_html

    def filter_series(self, series: List[Series]) -> List[Series]:
        filtered_series = []
        for s in series:
            description = (s.series_description or "").strip()
            instance_count = session.query(Instances).filter(Instances.series_id == s.id).count()

            # Check for exact match with "Loc"
            if description in ["Loc", "LOC"]:
                continue

            for line in [" MIP ", " MIP", "MIP "]:
                if line in description:
                    continue

            # Check for keywords in the description
            if any(
                keyword in description.lower()
                for keyword in ["apparent diffusion coefficient", "dynacad", "loc breast", "3 plane loc", "reformat"]
            ):
                continue

            # Ensure the series has at least 10 instances
            if instance_count < 10:
                continue

            filtered_series.append(s)

        return filtered_series

    def go_to_next_study(self):
        current_patient = self.get_patients()[self.current_patient_index]
        studies = session.query(Studies).filter_by(patient_id=current_patient.id).all()

        if self.current_study_index + 1 < len(studies):
            self.current_study_index += 1
            next_study_id = str(studies[self.current_study_index].id)
            self.study_select.value = next_study_id
            self.load_study_series(int(next_study_id))
        else:
            self.current_patient_index += 1
            self.current_study_index = 0

            if self.current_patient_index < len(self.get_patients()):
                next_patient = self.get_patients()[self.current_patient_index]
                self.patient_select.value = str(next_patient.id)
                self.load_patient_studies(next_patient.id)
            else:
                self.series_list_column.children = [Div(text="<h3>All patients have been labeled.</h3>")]

    def go_to_previous_study(self):
        if self.current_study_index > 0:
            self.current_study_index -= 1
            current_patient = self.get_patients()[self.current_patient_index]
            studies = session.query(Studies).filter_by(patient_id=current_patient.id).all()
            prev_study_id = str(studies[self.current_study_index].id)
            self.study_select.value = prev_study_id
            self.load_study_series(int(prev_study_id))
        elif self.current_patient_index > 0:
            self.current_patient_index -= 1
            current_patient = self.get_patients()[self.current_patient_index]
            studies = session.query(Studies).filter_by(patient_id=current_patient.id).all()
            if studies:
                self.current_study_index = len(studies) - 1
                prev_study_id = str(studies[self.current_study_index].id)
                self.patient_select.value = str(current_patient.id)
                self.study_select.value = prev_study_id
                self.load_study_series(int(prev_study_id))
        else:
            self.series_info_div.text = "<h3>At the beginning of the dataset.</h3>"

    def save_current_selection(self):
        try:
            for series_id, label_select, phase_select, _ in self.current_series_entries:
                selected_label = label_select.value.strip()
                if not selected_label:
                    continue

                selected_phase = (
                    phase_select.value.strip() if phase_select.visible and phase_select.value != "N/A" else None
                )
                series = session.query(Series).filter(Series.id == series_id).first()
                if series:
                    series.label = selected_label
                    if selected_label in self.label_manager.phase_labels and selected_phase:
                        series.phase = int(selected_phase)
                    else:
                        series.phase = None
                    self.label_manager.save_label(series, selected_label, selected_phase)

            session.commit()
            self.label_manager.labels_session.commit()
            self.go_to_next_study()

        except Exception as e:
            print(f"Error saving labels: {e}")
            session.rollback()
            self.label_manager.labels_session.rollback()


labeler = DicomSeriesLabeler()
doc = curdoc()
doc.add_root(labeler.layout)
doc.title = "DICOM Series Labeling"
