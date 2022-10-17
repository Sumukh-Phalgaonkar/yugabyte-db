// Copyright (c) YugaByte, Inc.
package com.yugabyte.yw.commissioner.tasks;

import com.yugabyte.yw.commissioner.BaseTaskDependencies;
import com.yugabyte.yw.commissioner.UserTaskDetails;
import com.yugabyte.yw.models.Universe;
import com.yugabyte.yw.models.XClusterConfig;
import com.yugabyte.yw.models.XClusterTableConfig;
import java.util.Set;
import javax.inject.Inject;
import lombok.extern.slf4j.Slf4j;

@Slf4j
public class RestartXClusterConfig extends CreateXClusterConfig {

  @Inject
  protected RestartXClusterConfig(BaseTaskDependencies baseTaskDependencies) {
    super(baseTaskDependencies);
  }

  @Override
  public void run() {
    log.info("Running {}", getName());

    XClusterConfig xClusterConfig = getXClusterConfigFromTaskParams();
    Universe sourceUniverse = Universe.getOrBadRequest(xClusterConfig.sourceUniverseUUID);
    Universe targetUniverse = Universe.getOrBadRequest(xClusterConfig.targetUniverseUUID);
    try {
      // Lock the source universe.
      lockUniverseForUpdate(sourceUniverse.universeUUID, sourceUniverse.version);
      try {
        // Lock the target universe.
        lockUniverseForUpdate(targetUniverse.universeUUID, targetUniverse.version);

        createXClusterConfigSetStatusTask(XClusterConfig.XClusterConfigStatusType.Updating)
            .setSubTaskGroupType(UserTaskDetails.SubTaskGroupType.DeleteXClusterReplication);

        // Set needBootstrap to true for all tables. It will check if it is required.
        xClusterConfig.setNeedBootstrapForTables(
            xClusterConfig.getTables(), true /* needBootstrap */);

        // Delete the xCluster config.
        createDeleteXClusterConfigSubtasks(
            xClusterConfig, true /* keepEntry */, taskParams().isForced());

        createXClusterConfigSetStatusTask(XClusterConfig.XClusterConfigStatusType.Updating);

        xClusterConfig.setStatusForTables(
            getTableIds(taskParams().getTableInfoList()), XClusterTableConfig.Status.Updating);

        addSubtasksToCreateXClusterConfig(
            sourceUniverse,
            targetUniverse,
            taskParams().getTableInfoList(),
            taskParams().getMainTableIndexTablesMap());

        createXClusterConfigSetStatusTask(XClusterConfig.XClusterConfigStatusType.Running)
            .setSubTaskGroupType(UserTaskDetails.SubTaskGroupType.ConfigureUniverse);

        createMarkUniverseUpdateSuccessTasks(targetUniverse.universeUUID)
            .setSubTaskGroupType(UserTaskDetails.SubTaskGroupType.ConfigureUniverse);

        createMarkUniverseUpdateSuccessTasks(sourceUniverse.universeUUID)
            .setSubTaskGroupType(UserTaskDetails.SubTaskGroupType.ConfigureUniverse);

        getRunnableTask().runSubTasks();
      } finally {
        // Unlock the target universe.
        unlockUniverseForUpdate(targetUniverse.universeUUID);
      }
    } catch (Exception e) {
      log.error("{} hit error : {}", getName(), e.getMessage());
      setXClusterConfigStatus(XClusterConfig.XClusterConfigStatusType.Failed);
      // Set tables in updating status to failed.
      Set<String> tablesInUpdatingStatus =
          xClusterConfig.getTableIdsInStatus(
              getTableIds(taskParams().getTableInfoList()), XClusterTableConfig.Status.Updating);
      xClusterConfig.setStatusForTables(tablesInUpdatingStatus, XClusterTableConfig.Status.Failed);
      throw new RuntimeException(e);
    } finally {
      // Unlock the source universe.
      unlockUniverseForUpdate(sourceUniverse.universeUUID);
    }

    log.info("Completed {}", getName());
  }
}
