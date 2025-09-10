// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
import React from "react";
import { User, Tenant, Study } from "../hooks/useAdminData";
import { StatCard } from "../../../components/ui/stat-card";
import {
  Users,
  Building2,
  FileText,
  Image,
  FolderOpen,
  UserCheck,
  KeyRound,
} from "lucide-react";

interface AdminStatsProps {
  users: User[];
  tenants: Tenant[];
  studies: Study[];
  slides: number;
  studiesCount: number;
  cases: number;
  loading?: boolean;
}

const AdminStats: React.FC<AdminStatsProps> = ({
  users,
  tenants,
  studies,
  studiesCount,
  slides,
  cases,
  loading = false,
}) => {
  // Ensure we have valid arrays
  const safeUsers = Array.isArray(users) ? users : [];
  const safeTenants = Array.isArray(tenants) ? tenants : [];
  const safeStudies = Array.isArray(studies) ? studies : [];

  const activeUsers = safeUsers.filter(
    (user) => user?.isActive === true
  ).length;
  const inactiveUsers = safeUsers.filter(
    (user) => user?.isActive === false
  ).length;
  const usersNeedingReset = safeUsers.filter(
    (user) => user?.mustResetPassword === true
  ).length;

  const publishedStudies = safeStudies.filter(
    (study) => study?.isPublished === true
  ).length;
  const draftStudies = safeStudies.filter(
    (study) => study?.isPublished === false
  ).length;

  if (loading) {
    return (
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-6">
        {[...Array(7)].map((_, i) => (
          <StatCard
            key={i}
            title="Loading..."
            value="--"
            className="animate-pulse"
          />
        ))}
      </div>
    );
  }

  const stats = [
    {
      title: "Total Users",
      value: safeUsers.length || 0,
      icon: Users,
      subtitle: `${activeUsers} active, ${inactiveUsers} inactive`,
    },
    {
      title: "Active Users",
      value: activeUsers || 0,
      icon: UserCheck,
      subtitle: "Currently active users",
    },
    {
      title: "Password Resets",
      value: usersNeedingReset || 0,
      icon: KeyRound,
      subtitle: "Users requiring password reset",
    },
    {
      title: "Total Tenants",
      value: safeTenants.length || 0,
      icon: Building2,
      subtitle: "Configured tenant organizations",
    },
    {
      title: "Total Studies",
      value: studiesCount || 0,
      icon: FileText,
      subtitle: `${publishedStudies} published, ${draftStudies} draft`,
    },
    {
      title: "Total Cases",
      value: cases || 0,
      icon: FolderOpen,
      subtitle: "Cases in the system",
    },
    {
      title: "Total Slides",
      value: slides || 0,
      icon: Image,
      subtitle: "Slides in the system",
    },
  ];

  return (
    <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-6">
      {stats.map((stat) => (
        <StatCard
          key={stat.title}
          title={stat.title}
          value={stat.value}
          subtitle={stat.subtitle}
          icon={stat.icon}
        />
      ))}
    </div>
  );
};

export default AdminStats;
