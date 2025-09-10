// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import { apiFetch } from "@/utils/fetchUtils";

interface ApiResponse<T> {
  data: T;
}

class ApiClient {
  private baseURL = "/api/v1";

  async get<T>(endpoint: string): Promise<ApiResponse<T>> {
    const data = await apiFetch<T>(`${this.baseURL}${endpoint}`);
    return { data };
  }

  async post<T>(endpoint: string, body?: any): Promise<ApiResponse<T>> {
    const data = await apiFetch<T>(`${this.baseURL}${endpoint}`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: body ? JSON.stringify(body) : undefined,
    });
    return { data };
  }

  async put<T>(endpoint: string, body?: any): Promise<ApiResponse<T>> {
    const data = await apiFetch<T>(`${this.baseURL}${endpoint}`, {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
      },
      body: body ? JSON.stringify(body) : undefined,
    });
    return { data };
  }

  async patch<T>(endpoint: string, body?: any): Promise<ApiResponse<T>> {
    const data = await apiFetch<T>(`${this.baseURL}${endpoint}`, {
      method: "PATCH",
      headers: {
        "Content-Type": "application/json",
      },
      body: body ? JSON.stringify(body) : undefined,
    });
    return { data };
  }

  async delete<T = void>(endpoint: string): Promise<ApiResponse<T>> {
    const data = await apiFetch<T>(`${this.baseURL}${endpoint}`, {
      method: "DELETE",
    });
    return { data };
  }

  async deleteWithBody<T = void>(
    endpoint: string,
    body?: any
  ): Promise<ApiResponse<T>> {
    const data = await apiFetch<T>(`${this.baseURL}${endpoint}`, {
      method: "DELETE",
      headers: {
        "Content-Type": "application/json",
      },
      body: body ? JSON.stringify(body) : undefined,
    });
    return { data };
  }
}

export const apiClient = new ApiClient();
