# Copyright 2025 AI for Oncology Research Group. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
import yaml
from typing import Optional

from lightning import LightningModule, Trainer
from typing_extensions import override
from eva.core.callbacks import config
from eva.core.loggers.log.parameters import log_parameters


class ConfigurationLogger(config.ConfigurationLogger):
    """
    Wrapper around ConfigurationLogger from EVA to avoid not logging configuration
    when logging directory cannot be found.

    This class extends EVA's ConfigurationLogger to handle MLFlowLogger's different
    directory structure.

    Parameters
    ----------
    verbose : bool, optional
        Whether to print configuration to console, by default True

    Notes
    -----
    MLFlowLogger uses a significantly different directory structure compared to
    the existing loggers within EVA. This wrapper ensures configuration logging
    works properly with MLFlowLogger.
    """

    def __init__(self, verbose: bool = True) -> None:
        """
        Inits :class:`ConfigurationLogger`.

        Parameters
        ----------
        verbose : bool, optional
            Whether to print configuration to console, by default True
        """
        super().__init__(verbose)

    @override
    def setup(self, trainer: Trainer, pl_module: LightningModule, stage: Optional[str] = None) -> None:
        """
        Called when the trainer setup begins. Loads the submitted configuration and logs it.

        Parameters
        ----------
        trainer : Trainer
            The trainer instance being used
        pl_module : LightningModule
            The LightningModule being trained
        stage : Optional[str], optional
            The stage of training being run (e.g. 'fit', 'validate', 'test'), by default None

        Returns
        -------
        None
            This method doesn't return anything

        Notes
        -----
        This method loads the configuration that was submitted to start the training run,
        prints it to console if verbose mode is enabled, and logs it to all configured
        loggers using the 'configuration' tag.
        """
        configuration = config._load_submitted_config()

        if self._verbose:
            config_as_text = yaml.dump(configuration, sort_keys=False)
            print(f"Configuration:\033[94m\n---\n{config_as_text}\033[0m")

        log_parameters(trainer.loggers, tag="configuration", parameters=configuration)
