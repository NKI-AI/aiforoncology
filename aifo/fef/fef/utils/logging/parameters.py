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
from typing import Any, Dict

from lightning.pytorch.loggers import MLFlowLogger
from eva.core.loggers import log_parameters


@log_parameters.register
def _(logger: MLFlowLogger, tag: str, parameters: Dict[str, Any]) -> None:
    """
    Adds parameters to an instance of MLFlowLogger.

    Parameters
    ----------
    logger : MLFlowLogger
        The MLFlow logger instance to add parameters to
    tag : str
        The tag to group the parameters under
    parameters : dict[str, Any]
        Dictionary of parameter names and values to log

    Returns
    -------
    None

    Notes
    -----
    This function gets automatically picked up by the existing `log_parameters` dispatcher.
    """
    logger.experiment.log_dict(dictionary=parameters, artifact_file=f"{tag}.yaml", run_id=logger.run_id)
