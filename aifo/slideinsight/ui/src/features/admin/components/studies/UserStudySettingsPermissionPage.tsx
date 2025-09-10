// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React, { useCallback, useMemo, useState } from "react";
import { useParams } from "@tanstack/react-router";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/AlertDialog";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Switch } from "@/components/ui/switch";
import {
  PlusIcon,
  TrashIcon,
  MagnifyingGlassIcon,
  ArrowPathIcon,
} from "@heroicons/react/24/outline";
import { useUsers, usePermissions as useAllPermissions } from "@/api/hooks";
import {
  useObjectGrants,
  useCreateObjectGrant,
  useDeleteObjectGrant,
} from "@/features/admin/hooks/useObjectGrants";

export default function UserStudySettingsPermissionPage() {
  const { studyUid } = useParams({
    from: "/_authenticated/studies/$studyUid/settings",
  });

  const {
    data: usersList = [],
    loading: usersLoading,
    refetch: refetchUsers,
  } = useUsers({ limit: 200 });
  const {
    data: permissionsList = [],
    loading: permissionsLoading,
    refetch: refetchPerms,
  } = useAllPermissions({
    page: 1,
    limit: 1000,
  });
  const {
    data: grants = [],
    isLoading: grantsLoading,
    refetch: refetchGrants,
  } = useObjectGrants("study", studyUid);

  const createGrant = useCreateObjectGrant();
  const deleteGrant = useDeleteObjectGrant();

  // Dialog state
  const [showAddPermission, setShowAddPermission] = useState(false);
  const [selectedUserUid, setSelectedUserUid] = useState("");
  const [selectedPermission, setSelectedPermission] = useState("");
  const [confirmRevokeUserKey, setConfirmRevokeUserKey] = useState<
    string | null
  >(null);

  // Toolbar state
  const [query, setQuery] = useState("");
  const [permFilter, setPermFilter] = useState<string>("__all__");
  const [onlyWithAccess, setOnlyWithAccess] = useState(false);

  // Narrow to study.* permissions
  const studyPerms = useMemo(
    () => permissionsList.filter((p: any) => p.name?.startsWith("studies.")),
    [permissionsList]
  );

  // Group grants by user
  type UserRow = {
    key: string;
    grantee_type: string;
    grantee_id: number;
    name: string;
    email?: string;
    perms: Set<string>;
  };

  const groupedByUser: UserRow[] = useMemo(() => {
    const byUser = new Map<string, UserRow>();
    grants.forEach((g: any) => {
      const key = `${g.grantee_type}:${g.grantee_id}`;
      if (!byUser.has(key)) {
        byUser.set(key, {
          key,
          grantee_type: g.grantee_type,
          grantee_id: g.grantee_id,
          name:
            g.grantee_info?.name ||
            g.grantee_name ||
            `${g.grantee_type} ${g.grantee_id}`,
          email: g.grantee_info?.email,
          perms: new Set<string>(),
        });
      }
      byUser.get(key)!.perms.add(g.permission);
    });
    return Array.from(byUser.values()).sort((a, b) =>
      a.name.localeCompare(b.name)
    );
  }, [grants]);

  // Filtered view
  const filteredRows = useMemo(() => {
    const q = query.trim().toLowerCase();
    return groupedByUser.filter((row) => {
      const matchesQuery =
        !q ||
        row.name.toLowerCase().includes(q) ||
        (row.email ? row.email.toLowerCase().includes(q) : false);

      const matchesPerm =
        permFilter === "__all__" ? true : row.perms.has(permFilter);
      const matchesHasAccess = onlyWithAccess ? row.perms.size > 0 : true;
      return matchesQuery && matchesPerm && matchesHasAccess;
    });
  }, [groupedByUser, query, permFilter, onlyWithAccess]);

  const anyMutationPending = createGrant.isPending || deleteGrant.isPending;

  const togglePermission = useCallback(
    async (row: UserRow, permName: string, next: boolean) => {
      try {
        if (next) {
          await createGrant.mutateAsync({
            grantee_type: row.grantee_type as any,
            grantee_id: row.grantee_id,
            permission: permName,
            resource_type: "study",
            resource_uid: studyUid,
          });
        } else {
          await deleteGrant.mutateAsync({
            resourceType: "study",
            resourceId: studyUid,
            data: {
              grantee_type: row.grantee_type as any,
              grantee_id: row.grantee_id,
              permission: permName,
            },
          });
        }
      } finally {
        // Keep UI in sync
        void refetchGrants?.();
      }
    },
    [createGrant, deleteGrant, studyUid, refetchGrants]
  );

  const revokeAllForUser = useCallback(
    async (row: UserRow) => {
      const revoke = Array.from(row.perms).filter((n: string) =>
        n.startsWith("studies.")
      );
      for (const perm of revoke) {
        await deleteGrant.mutateAsync({
          resourceType: "study",
          resourceId: studyUid,
          data: {
            grantee_type: row.grantee_type as any,
            grantee_id: row.grantee_id,
            permission: perm,
          },
        });
      }
      void refetchGrants?.();
    },
    [deleteGrant, studyUid, refetchGrants]
  );

  const uid = React.useId();

  return (
    <Card className="border bg-card shadow-sm">
      <CardHeader className="pb-3">
        <CardTitle className="flex items-center justify-between">
          <div>
            <div className="text-base font-semibold">Permissions</div>
            <p className="mt-1 text-sm text-muted-foreground">
              Grant and manage access to this study. Toggle granular
              capabilities per user.
            </p>
          </div>
          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => {
                void refetchUsers?.();
                void refetchPerms?.();
                void refetchGrants?.();
              }}
            >
              <ArrowPathIcon className="mr-2 h-4 w-4" />
              Refresh
            </Button>
            <Button size="sm" onClick={() => setShowAddPermission(true)}>
              <PlusIcon className="mr-2 h-4 w-4" />
              Add permission
            </Button>
          </div>
        </CardTitle>
      </CardHeader>

      <CardContent className="space-y-4">
        {/* Toolbar */}
        <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div className="relative w-full sm:max-w-xs">
            <MagnifyingGlassIcon className="pointer-events-none absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
            <Input
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder="Search user or email…"
              className="pl-8"
              aria-label="Search grantees"
            />
          </div>

          <div className="flex flex-wrap items-center gap-3">
            <div className="flex items-center gap-2">
              <Label
                htmlFor={`${uid}-only-access`}
                className="text-sm text-muted-foreground"
              >
                Only users with access
              </Label>
              <Switch
                id={`${uid}-only-access`}
                checked={onlyWithAccess}
                onCheckedChange={(v) => setOnlyWithAccess(v === true)}
              />
            </div>

            <Select value={permFilter} onValueChange={setPermFilter}>
              <SelectTrigger className="w-[220px]">
                <SelectValue placeholder="Filter by permission" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="__all__">All permissions</SelectItem>
                {studyPerms.map((p: any) => (
                  <SelectItem key={p.name} value={p.name}>
                    {p.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        </div>

        {/* Table */}
        <div className="overflow-hidden rounded-lg border">
          <div className="max-h-[60vh] overflow-auto">
            <Table>
              <TableHeader className="sticky top-0 z-10 bg-muted/40 backdrop-blur">
                <TableRow>
                  <TableHead className="w-[35%]">User</TableHead>
                  <TableHead className="w-[10%]">Type</TableHead>
                  <TableHead className="w-[45%]">Permissions</TableHead>
                  <TableHead className="w-[10%] text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {grantsLoading ? (
                  Array.from({ length: 5 }).map((_, i) => (
                    <TableRow key={`skeleton-${i}`}>
                      <TableCell className="py-4">
                        <div className="h-4 w-40 animate-pulse rounded bg-muted" />
                        <div className="mt-2 h-3 w-24 animate-pulse rounded bg-muted" />
                      </TableCell>
                      <TableCell>
                        <div className="h-4 w-16 animate-pulse rounded bg-muted" />
                      </TableCell>
                      <TableCell>
                        <div className="flex flex-wrap gap-3">
                          <div className="h-5 w-28 animate-pulse rounded bg-muted" />
                          <div className="h-5 w-24 animate-pulse rounded bg-muted" />
                          <div className="h-5 w-32 animate-pulse rounded bg-muted" />
                        </div>
                      </TableCell>
                      <TableCell className="text-right">
                        <div className="ml-auto h-5 w-8 animate-pulse rounded bg-muted" />
                      </TableCell>
                    </TableRow>
                  ))
                ) : filteredRows.length === 0 ? (
                  <TableRow>
                    <TableCell
                      colSpan={4}
                      className="py-10 text-center text-sm text-muted-foreground"
                    >
                      {groupedByUser.length === 0
                        ? "No permissions yet. Use “Add permission” to grant access."
                        : "No results. Adjust your search or filters."}
                    </TableCell>
                  </TableRow>
                ) : (
                  filteredRows.map((row) => (
                    <TableRow key={row.key} className="align-top">
                      {/* User */}
                      <TableCell>
                        <div className="flex flex-col">
                          <span className="font-medium">{row.name}</span>
                          {row.email && (
                            <span className="text-xs text-muted-foreground">
                              {row.email}
                            </span>
                          )}
                        </div>
                      </TableCell>

                      {/* Type */}
                      <TableCell className="capitalize">
                        {row.grantee_type}
                      </TableCell>

                      {/* Permissions */}
                      <TableCell>
                        <div className="flex flex-wrap gap-2">
                          {studyPerms.map((p: any) => {
                            const checked = row.perms.has(p.name);
                            const id = `${row.key}-${p.name}`;
                            const disabled = anyMutationPending;
                            return (
                              <label
                                key={p.name}
                                htmlFor={id}
                                className={[
                                  "inline-flex select-none items-center gap-2 rounded-md border px-2 py-1",
                                  checked
                                    ? "border-primary/30 bg-primary/5"
                                    : "border-muted-foreground/20 bg-transparent hover:bg-muted/30",
                                  disabled ? "opacity-50" : "",
                                  "text-xs",
                                ].join(" ")}
                              >
                                <Checkbox
                                  id={id}
                                  checked={checked}
                                  disabled={disabled}
                                  onCheckedChange={(v) =>
                                    togglePermission(row, p.name, v === true)
                                  }
                                />
                                <code className="font-mono text-[11px]">
                                  {p.name}
                                </code>
                              </label>
                            );
                          })}
                        </div>
                      </TableCell>

                      {/* Actions */}
                      <TableCell className="text-right">
                        <Button
                          variant="ghost"
                          size="icon"
                          className="text-destructive hover:bg-destructive/10 hover:text-destructive/80"
                          onClick={() => setConfirmRevokeUserKey(row.key)}
                          aria-label={`Revoke all permissions for ${row.name}`}
                        >
                          <TrashIcon className="h-4 w-4" />
                        </Button>
                      </TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </div>
        </div>

        {/* Footnote */}
        <p className="text-xs text-muted-foreground">
          Changes apply immediately. Use the refresh button if you’re managing
          permissions in multiple tabs.
        </p>
      </CardContent>

      {/* Add Permission Dialog */}
      <AlertDialog open={showAddPermission} onOpenChange={setShowAddPermission}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Add permission</AlertDialogTitle>
            <AlertDialogDescription>
              Grant a study permission to a user.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <div className="space-y-3 py-2">
            <div className="space-y-1">
              <Label>User</Label>
              <Select
                value={selectedUserUid}
                onValueChange={setSelectedUserUid}
              >
                <SelectTrigger>
                  <SelectValue
                    placeholder={
                      usersLoading ? "Loading users…" : "Choose a user"
                    }
                  />
                </SelectTrigger>
                <SelectContent>
                  {usersList.map((u) => (
                    <SelectItem key={u.userUid} value={u.userUid}>
                      <div className="flex items-center gap-2">
                        <span className="text-sm font-medium">
                          {u.firstName} {u.lastName}
                        </span>
                        <span className="text-xs text-muted-foreground">
                          {u.email}
                        </span>
                      </div>
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-1">
              <Label>Permission</Label>
              <Select
                value={selectedPermission}
                onValueChange={setSelectedPermission}
              >
                <SelectTrigger>
                  <SelectValue
                    placeholder={
                      permissionsLoading ? "Loading…" : "Choose permission"
                    }
                  />
                </SelectTrigger>
                <SelectContent>
                  {studyPerms.map((p: any) => (
                    <SelectItem key={p.name} value={p.name}>
                      {p.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>
          <AlertDialogFooter>
            <AlertDialogCancel onClick={() => setShowAddPermission(false)}>
              Cancel
            </AlertDialogCancel>
            <AlertDialogAction
              onClick={async () => {
                if (!selectedUserUid || !selectedPermission) return;
                await createGrant.mutateAsync({
                  grantee_type: "user",
                  grantee_uid: selectedUserUid,
                  permission: selectedPermission,
                  resource_type: "study",
                  resource_uid: studyUid,
                });
                setShowAddPermission(false);
                setSelectedUserUid("");
                setSelectedPermission("");
                void refetchGrants?.();
              }}
              disabled={
                !selectedUserUid || !selectedPermission || createGrant.isPending
              }
            >
              {createGrant.isPending ? "Granting…" : "Grant"}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      {/* Confirm Revoke All */}
      <AlertDialog
        open={!!confirmRevokeUserKey}
        onOpenChange={(o) => !o && setConfirmRevokeUserKey(null)}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Revoke all permissions?</AlertDialogTitle>
            <AlertDialogDescription>
              This will remove all <code className="font-mono">studies.*</code>{" "}
              permissions for the selected user.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel onClick={() => setConfirmRevokeUserKey(null)}>
              Cancel
            </AlertDialogCancel>
            <AlertDialogAction
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
              onClick={async () => {
                const row = groupedByUser.find(
                  (r) => r.key === confirmRevokeUserKey
                );
                if (row) await revokeAllForUser(row);
                setConfirmRevokeUserKey(null);
              }}
              disabled={deleteGrant.isPending}
            >
              {deleteGrant.isPending ? "Revoking…" : "Revoke"}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </Card>
  );
}
