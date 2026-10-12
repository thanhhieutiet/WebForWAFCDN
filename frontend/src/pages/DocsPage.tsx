import React, { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import {
  FileCode,
  Search,
  Tag,
  ChevronDown,
  ChevronRight,
} from 'lucide-react';
import { api } from '@/lib/api';
import { Input } from '@/components/ui/input';
import { Skeleton } from '@/components/ui/skeleton';
import { CurlBlock } from '@/components/lab/CurlBlock';

export function DocsPage(): React.JSX.Element {
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedTag, setSelectedTag] = useState<string>('all');
  const [expandedEndpoints, setExpandedEndpoints] = useState<Record<string, boolean>>({});

  // Query OpenAPI specification JSON
  const { data: spec, isLoading, isError } = useQuery({
    queryKey: ['openapi-spec'],
    queryFn: async () => {
      const res = await api.get<any>('/openapi.json');
      return res.data;
    },
  });

  const toggleEndpoint = (key: string) => {
    setExpandedEndpoints((prev) => ({ ...prev, [key]: !prev[key] }));
  };

  const paths = spec?.paths || {};
  const endpointList: Array<{
    path: string;
    method: string;
    operation: any;
    key: string;
  }> = [];

  Object.entries(paths).forEach(([pathKey, pathItem]: [string, any]) => {
    ['get', 'post', 'put', 'patch', 'delete'].forEach((method) => {
      if (pathItem[method]) {
        endpointList.push({
          path: pathKey,
          method: method.toUpperCase(),
          operation: pathItem[method],
          key: `${method.toUpperCase()}_${pathKey}`,
        });
      }
    });
  });

  const allTags = Array.from(
    new Set(
      endpointList.flatMap((e) => e.operation.tags || ['General'])
    )
  ).sort();

  const filteredEndpoints = endpointList.filter((ep) => {
    const matchesSearch =
      ep.path.toLowerCase().includes(searchQuery.toLowerCase()) ||
      (ep.operation.summary &&
        ep.operation.summary.toLowerCase().includes(searchQuery.toLowerCase()));

    const matchesTag =
      selectedTag === 'all' ||
      (ep.operation.tags && ep.operation.tags.includes(selectedTag));

    return matchesSearch && matchesTag;
  });

  const getMethodBadgeClass = (method: string) => {
    switch (method) {
      case 'GET':
        return 'bg-blue-100 text-blue-800 dark:bg-blue-950 dark:text-blue-300 border-blue-200 dark:border-blue-900';
      case 'POST':
        return 'bg-teal-100 text-teal-800 dark:bg-teal-950 dark:text-teal-300 border-teal-200 dark:border-teal-900';
      case 'PUT':
        return 'bg-amber-100 text-amber-800 dark:bg-amber-950 dark:text-amber-300 border-amber-200 dark:border-amber-900';
      case 'PATCH':
        return 'bg-purple-100 text-purple-800 dark:bg-purple-950 dark:text-purple-300 border-purple-200 dark:border-purple-900';
      case 'DELETE':
        return 'bg-rose-100 text-rose-800 dark:bg-rose-950 dark:text-rose-300 border-rose-200 dark:border-rose-900';
      default:
        return 'bg-slate-100 text-slate-800';
    }
  };

  return (
    <div className="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8 space-y-8">
      {/* Header */}
      <div className="space-y-3 pb-6 border-b border-slate-200 dark:border-slate-800">
        <div className="inline-flex items-center gap-2 rounded-control bg-teal-50 px-3 py-1 text-xs font-semibold text-teal-700 ring-1 ring-inset ring-teal-600/20 dark:bg-teal-950/60 dark:text-teal-300">
          <FileCode className="h-3.5 w-3.5" />
          <span>OpenAPI 3.0.3 Specification</span>
        </div>

        <h1 className="text-3xl font-bold tracking-tight text-slate-900 dark:text-white">
          Tài liệu Đặc tả API (Origin API Docs)
        </h1>
        <p className="text-sm text-slate-600 dark:text-slate-300 max-w-3xl leading-relaxed">
          Được trích xuất tự động từ <code className="rounded bg-slate-100 dark:bg-slate-800 px-1 py-0.5 font-mono text-xs">/openapi.json</code> của máy chủ Go. Tự render hoàn toàn trên client mà không sử dụng các CDN script bên ngoài.
        </p>
      </div>

      {/* Controls Bar: Search & Tag Filter */}
      <div className="flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-4">
        <div className="relative flex-1 max-w-md">
          <Input
            type="search"
            placeholder="Tìm theo URL path hoặc mô tả endpoint..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-9 text-xs"
          />
          <Search className="absolute left-3 top-3 h-4 w-4 text-slate-400 pointer-events-none" />
        </div>

        <div className="flex flex-wrap items-center gap-1.5 text-xs">
          <span className="text-slate-400 mr-1 flex items-center gap-1">
            <Tag className="h-3 w-3" />
            Nhóm:
          </span>
          <button
            type="button"
            onClick={() => setSelectedTag('all')}
            className={`rounded-control px-2.5 py-1 font-medium transition-colors ${
              selectedTag === 'all'
                ? 'bg-teal-600 text-white font-semibold'
                : 'bg-slate-100 text-slate-600 hover:bg-slate-200 dark:bg-slate-800 dark:text-slate-300'
            }`}
          >
            Tất cả
          </button>
          {allTags.map((tag) => (
            <button
              key={tag}
              type="button"
              onClick={() => setSelectedTag(tag)}
              className={`rounded-control px-2.5 py-1 font-medium transition-colors ${
                selectedTag === tag
                  ? 'bg-teal-600 text-white font-semibold'
                  : 'bg-slate-100 text-slate-600 hover:bg-slate-200 dark:bg-slate-800 dark:text-slate-300'
              }`}
            >
              {tag}
            </button>
          ))}
        </div>
      </div>

      {isLoading && (
        <div className="space-y-4">
          <Skeleton className="h-16 w-full rounded-card" />
          <Skeleton className="h-16 w-full rounded-card" />
          <Skeleton className="h-16 w-full rounded-card" />
        </div>
      )}

      {isError && (
        <div className="rounded-card border border-rose-200 bg-rose-50 p-6 text-center text-xs text-rose-700 dark:bg-rose-950/20">
          Không thể tải tài liệu đặc tả /openapi.json từ origin.
        </div>
      )}

      {/* Endpoints Accordion List */}
      <div className="space-y-3">
        {filteredEndpoints.map((ep) => {
          const isExpanded = !!expandedEndpoints[ep.key];
          const params = ep.operation.parameters || [];
          const responses = ep.operation.responses || {};

          return (
            <div
              key={ep.key}
              className="rounded-card border border-slate-200 bg-white overflow-hidden shadow-sm dark:border-slate-800 dark:bg-slate-900"
            >
              <div
                onClick={() => toggleEndpoint(ep.key)}
                className="flex items-center justify-between p-4 cursor-pointer hover:bg-slate-50 dark:hover:bg-slate-800/40 transition-colors"
              >
                <div className="flex items-center gap-3 overflow-hidden">
                  <span
                    className={`font-mono font-bold text-xs px-2.5 py-1 rounded border shrink-0 ${getMethodBadgeClass(
                      ep.method
                    )}`}
                  >
                    {ep.method}
                  </span>
                  <span className="font-mono text-sm font-semibold text-slate-900 dark:text-white truncate">
                    {ep.path}
                  </span>
                  <span className="hidden md:inline text-xs text-slate-500 truncate">
                    {ep.operation.summary}
                  </span>
                </div>

                <div className="flex items-center gap-2 shrink-0">
                  {ep.operation.tags?.[0] && (
                    <span className="hidden sm:inline rounded bg-slate-100 dark:bg-slate-800 px-2 py-0.5 text-[11px] text-slate-600 dark:text-slate-400 font-mono">
                      {ep.operation.tags[0]}
                    </span>
                  )}
                  {isExpanded ? (
                    <ChevronDown className="h-4 w-4 text-slate-400" />
                  ) : (
                    <ChevronRight className="h-4 w-4 text-slate-400" />
                  )}
                </div>
              </div>

              {isExpanded && (
                <div className="p-4 border-t border-slate-100 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-950/40 space-y-4 text-xs">
                  {ep.operation.description && (
                    <p className="text-slate-600 dark:text-slate-300 leading-relaxed">
                      {ep.operation.description}
                    </p>
                  )}

                  {/* cURL Example */}
                  <CurlBlock
                    curl={`curl -i -X ${ep.method} "https://shop.labshop.local${ep.path}"`}
                    label="Lệnh cURL gọi endpoint này"
                  />

                  {/* Parameters Table */}
                  {params.length > 0 && (
                    <div className="space-y-1.5">
                      <div className="font-semibold text-slate-700 dark:text-slate-300">
                        Tham số (Parameters):
                      </div>
                      <div className="rounded border border-slate-200 dark:border-slate-800 overflow-hidden font-mono text-[11px]">
                        <table className="w-full text-left">
                          <thead className="bg-slate-100 dark:bg-slate-800 text-slate-600">
                            <tr>
                              <th className="p-2">Name</th>
                              <th className="p-2">In</th>
                              <th className="p-2">Type</th>
                              <th className="p-2">Required</th>
                              <th className="p-2">Mô tả</th>
                            </tr>
                          </thead>
                          <tbody className="divide-y divide-slate-100 dark:divide-slate-800">
                            {params.map((p: any, idx: number) => (
                              <tr key={idx}>
                                <td className="p-2 font-bold text-teal-600">{p.name}</td>
                                <td className="p-2 text-slate-500">{p.in}</td>
                                <td className="p-2 text-slate-500">{p.schema?.type || 'string'}</td>
                                <td className="p-2 text-slate-500">{p.required ? 'Yes' : 'No'}</td>
                                <td className="p-2 font-sans text-slate-700 dark:text-slate-300">
                                  {p.description || '-'}
                                </td>
                              </tr>
                            ))}
                          </tbody>
                        </table>
                      </div>
                    </div>
                  )}

                  {/* Responses Table */}
                  <div className="space-y-1.5">
                    <div className="font-semibold text-slate-700 dark:text-slate-300">
                      Mã phản hồi (Responses):
                    </div>
                    <div className="space-y-1">
                      {Object.entries(responses).map(([code, resp]: [string, any]) => (
                        <div key={code} className="flex items-baseline gap-2 font-mono text-[11px]">
                          <span
                            className={`font-bold px-1.5 py-0.2 rounded ${
                              code.startsWith('2')
                                ? 'bg-teal-100 text-teal-800 dark:bg-teal-950 dark:text-teal-300'
                                : 'bg-rose-100 text-rose-800 dark:bg-rose-950 dark:text-rose-300'
                            }`}
                          >
                            {code}
                          </span>
                          <span className="font-sans text-slate-600 dark:text-slate-300">
                            {resp.description || 'Response'}
                          </span>
                        </div>
                      ))}
                    </div>
                  </div>
                </div>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
}
