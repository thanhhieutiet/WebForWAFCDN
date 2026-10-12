import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import {
  Headphones,
  Plus,
  Send,
  Paperclip,
  FileText,
  AlertCircle,
  User,
  Shield,
} from 'lucide-react';
import { api, ApiError } from '@/lib/api';
import { useAuth } from '@/lib/authContext';
import { formatDateTime } from '@/lib/formatters';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Dialog } from '@/components/ui/dialog';
import { Skeleton } from '@/components/ui/skeleton';
import { toast } from 'sonner';
import { getErrorMessage } from '@/lib/messages';

interface TicketMessageDTO {
  id: string;
  sender_id: string;
  sender_role: string;
  message: string;
  created_at: string;
}

interface TicketAttachmentDTO {
  id: string;
  file_name: string;
  file_size: number;
  content_type: string;
}

interface TicketDTO {
  id: string;
  user_id: string;
  user_email: string;
  subject: string;
  status: 'OPEN' | 'RESOLVED' | 'CLOSED';
  messages: TicketMessageDTO[];
  attachments: TicketAttachmentDTO[];
  created_at: string;
  updated_at: string;
}

export function SupportPage(): React.JSX.Element {
  const { isAuthenticated } = useAuth();
  const queryClient = useQueryClient();

  const [selectedTicketId, setSelectedTicketId] = useState<string | null>(null);
  const [createModalOpen, setCreateModalOpen] = useState(false);
  const [replyMessage, setReplyMessage] = useState('');

  // Create ticket state
  const [newSubject, setNewSubject] = useState('');
  const [newInitialMsg, setNewInitialMsg] = useState('');
  const [createError, setCreateError] = useState<string | null>(null);

  // File upload state
  const [uploading, setUploading] = useState(false);

  // Query user tickets
  const { data: ticketsData, isLoading } = useQuery({
    queryKey: ['my-tickets'],
    queryFn: async () => {
      const res = await api.get<{ items: TicketDTO[]; total: number }>('/api/support/tickets');
      return res.data;
    },
    enabled: isAuthenticated,
  });

  const tickets = ticketsData?.items || [];

  // Query selected ticket detail
  const { data: selectedTicket } = useQuery({
    queryKey: ['ticket-detail', selectedTicketId],
    queryFn: async () => {
      const res = await api.get<TicketDTO>(`/api/support/tickets/${selectedTicketId}`);
      return res.data;
    },
    enabled: !!selectedTicketId,
  });

  // Create Ticket Mutation
  const createMutation = useMutation({
    mutationFn: async () => {
      setCreateError(null);
      const res = await api.post<TicketDTO>('/api/support/tickets', {
        subject: newSubject.trim(),
        initial_message: newInitialMsg.trim(),
      });
      return res.data;
    },
    onSuccess: (created) => {
      toast.success('Đã gửi yêu cầu hỗ trợ mới thành công!');
      setCreateModalOpen(false);
      setNewSubject('');
      setNewInitialMsg('');
      setSelectedTicketId(created.id);
      queryClient.invalidateQueries({ queryKey: ['my-tickets'] });
    },
    onError: (err: any) => {
      const msg = err instanceof ApiError ? err.message : getErrorMessage('unknown_error');
      setCreateError(msg);
      toast.error('Gửi yêu cầu thất bại', { description: msg });
    },
  });

  // Send Reply Message Mutation
  const replyMutation = useMutation({
    mutationFn: async () => {
      const res = await api.post(`/api/support/tickets/${selectedTicketId}/messages`, {
        message: replyMessage.trim(),
      });
      return res.data;
    },
    onSuccess: () => {
      setReplyMessage('');
      queryClient.invalidateQueries({ queryKey: ['ticket-detail', selectedTicketId] });
      queryClient.invalidateQueries({ queryKey: ['my-tickets'] });
      toast.success('Đã gửi phản hồi');
    },
    onError: (err: any) => {
      const msg = err instanceof ApiError ? err.message : getErrorMessage('unknown_error');
      toast.error('Gửi phản hồi thất bại', { description: msg });
    },
  });

  // Upload Attachment
  const handleFileUpload = async (file: File) => {
    if (!selectedTicketId) return;

    // Validate size (max 10MB)
    if (file.size > 10 * 1024 * 1024) {
      toast.error('Kích thước tệp vượt quá 10MB cho phép');
      return;
    }

    const formData = new FormData();
    formData.append('file', file);

    try {
      setUploading(true);
      await api.post(`/api/support/tickets/${selectedTicketId}/attachments`, formData);
      toast.success(`Đã tải lên tệp: ${file.name}`);
      queryClient.invalidateQueries({ queryKey: ['ticket-detail', selectedTicketId] });
    } catch (err: any) {
      const msg = err instanceof ApiError ? err.message : 'Tải lên thất bại';
      toast.error('Không thể đính kèm tệp', { description: msg });
    } finally {
      setUploading(false);
    }
  };

  if (!isAuthenticated) {
    return (
      <div className="mx-auto max-w-xl px-4 py-16 text-center space-y-4">
        <Headphones className="mx-auto h-12 w-12 text-slate-400" />
        <h1 className="text-2xl font-bold">Trung tâm Hỗ trợ Kỹ thuật</h1>
        <p className="text-sm text-slate-500">
          Vui lòng đăng nhập để tạo ticket và trao đổi với đội ngũ kỹ thuật.
        </p>
        <Link to="/login">
          <Button>Đăng nhập ngay</Button>
        </Link>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8 space-y-6">
      {/* Top Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-4 border-b border-slate-200 dark:border-slate-800">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-slate-900 dark:text-white">
            Hỗ trợ Kỹ thuật & Khách hàng
          </h1>
          <p className="text-xs text-slate-500 dark:text-slate-400 mt-1">
            Gửi yêu cầu trợ giúp và theo dõi phản hồi trực tiếp từ đội ngũ hỗ trợ LabShop
          </p>
        </div>

        <Button onClick={() => setCreateModalOpen(true)} className="gap-2">
          <Plus className="h-4 w-4" />
          <span>Tạo yêu cầu mới</span>
        </Button>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8 items-start">
        {/* Left: Tickets List (1 Col) */}
        <div className="space-y-4">
          <div className="rounded-card border border-slate-200 bg-white dark:border-slate-800 dark:bg-slate-900 overflow-hidden shadow-sm">
            <div className="p-3 border-b border-slate-100 dark:border-slate-800 font-semibold text-xs text-slate-700 dark:text-slate-300">
              Yêu cầu của bạn ({tickets.length})
            </div>

            {isLoading ? (
              <div className="p-4 space-y-3">
                <Skeleton className="h-16 w-full rounded" />
                <Skeleton className="h-16 w-full rounded" />
              </div>
            ) : tickets.length === 0 ? (
              <div className="p-8 text-center text-xs text-slate-500 italic">
                Chưa có yêu cầu hỗ trợ nào.
              </div>
            ) : (
              <div className="divide-y divide-slate-100 dark:divide-slate-800 max-h-[550px] overflow-y-auto">
                {tickets.map((t) => {
                  const isSelected = t.id === selectedTicketId;
                  return (
                    <div
                      key={t.id}
                      onClick={() => setSelectedTicketId(t.id)}
                      className={`p-3.5 cursor-pointer text-xs transition-colors ${
                        isSelected
                          ? 'bg-teal-50/60 dark:bg-teal-950/40 border-l-2 border-teal-600'
                          : 'hover:bg-slate-50 dark:hover:bg-slate-800/50'
                      }`}
                    >
                      <div className="flex items-center justify-between gap-2 mb-1">
                        <span className="font-bold text-slate-900 dark:text-white line-clamp-1">
                          {t.subject}
                        </span>
                        <span
                          className={`rounded px-1.5 py-0.2 text-[10px] font-bold ${
                            t.status === 'OPEN'
                              ? 'bg-amber-100 text-amber-800 dark:bg-amber-950 dark:text-amber-300'
                              : t.status === 'RESOLVED'
                              ? 'bg-teal-100 text-teal-800 dark:bg-teal-950 dark:text-teal-300'
                              : 'bg-slate-100 text-slate-800 dark:bg-slate-800 dark:text-slate-300'
                          }`}
                        >
                          {t.status}
                        </span>
                      </div>
                      <div className="text-[11px] text-slate-400 font-mono">
                        #{t.id.slice(0, 8)} • {formatDateTime(t.created_at)}
                      </div>
                    </div>
                  );
                })}
              </div>
            )}
          </div>
        </div>

        {/* Right: Message Thread & Chat (2 Cols) */}
        <div className="lg:col-span-2">
          {selectedTicket ? (
            <div className="rounded-card border border-slate-200 bg-white dark:border-slate-800 dark:bg-slate-900 shadow-sm flex flex-col h-[650px] overflow-hidden">
              {/* Ticket header */}
              <div className="p-4 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between gap-4">
                <div>
                  <h2 className="text-base font-bold text-slate-900 dark:text-white">
                    {selectedTicket.subject}
                  </h2>
                  <div className="text-xs text-slate-500 font-mono mt-0.5">
                    Mã ticket: #{selectedTicket.id} • Khởi tạo lúc {formatDateTime(selectedTicket.created_at)}
                  </div>
                </div>

                <span
                  className={`rounded-control px-2.5 py-1 text-xs font-bold ${
                    selectedTicket.status === 'OPEN'
                      ? 'bg-amber-100 text-amber-800 dark:bg-amber-950 dark:text-amber-300'
                      : selectedTicket.status === 'RESOLVED'
                      ? 'bg-teal-100 text-teal-800 dark:bg-teal-950 dark:text-teal-300'
                      : 'bg-slate-100 text-slate-800'
                  }`}
                >
                  {selectedTicket.status}
                </span>
              </div>

              {/* Messages thread */}
              <div className="flex-1 p-4 overflow-y-auto space-y-4 bg-slate-50/50 dark:bg-slate-950/30">
                {selectedTicket.messages.map((m) => {
                  const isStaff = m.sender_role === 'admin' || m.sender_role === 'staff';

                  return (
                    <div
                      key={m.id}
                      className={`flex gap-3 max-w-[85%] ${
                        isStaff ? 'mr-auto' : 'ml-auto flex-row-reverse'
                      }`}
                    >
                      <div
                        className={`flex h-8 w-8 shrink-0 items-center justify-center rounded-full text-xs font-bold ${
                          isStaff
                            ? 'bg-teal-600 text-white'
                            : 'bg-slate-200 text-slate-800 dark:bg-slate-700 dark:text-slate-200'
                        }`}
                      >
                        {isStaff ? <Shield className="h-4 w-4" /> : <User className="h-4 w-4" />}
                      </div>

                      <div className="space-y-1">
                        <div
                          className={`text-[10px] text-slate-400 ${
                            isStaff ? 'text-left' : 'text-right'
                          }`}
                        >
                          {isStaff ? 'Đội ngũ Hỗ trợ' : 'Bạn'} • {formatDateTime(m.created_at)}
                        </div>

                        <div
                          className={`rounded-card p-3 text-xs leading-relaxed ${
                            isStaff
                              ? 'bg-white text-slate-800 shadow-sm border border-slate-200 dark:bg-slate-900 dark:text-slate-200 dark:border-slate-800'
                              : 'bg-teal-600 text-white shadow-sm'
                          }`}
                        >
                          {m.message}
                        </div>
                      </div>
                    </div>
                  );
                })}

                {/* Attachments list if any */}
                {selectedTicket.attachments && selectedTicket.attachments.length > 0 && (
                  <div className="pt-2 border-t border-slate-200 dark:border-slate-800 space-y-1.5">
                    <span className="text-[11px] font-semibold text-slate-500">
                      Tệp đính kèm ({selectedTicket.attachments.length}):
                    </span>
                    <div className="flex flex-wrap gap-2">
                      {selectedTicket.attachments.map((att) => (
                        <div
                          key={att.id}
                          className="inline-flex items-center gap-1.5 rounded border border-slate-200 bg-white px-2.5 py-1 text-xs text-slate-700 shadow-sm dark:border-slate-800 dark:bg-slate-900 dark:text-slate-300"
                        >
                          <FileText className="h-3.5 w-3.5 text-teal-600" />
                          <span className="font-mono text-[11px]">{att.file_name}</span>
                          <span className="text-[10px] text-slate-400">
                            ({Math.round(att.file_size / 1024)} KB)
                          </span>
                        </div>
                      ))}
                    </div>
                  </div>
                )}
              </div>

              {/* Reply box & Upload section */}
              {selectedTicket.status !== 'CLOSED' ? (
                <div className="p-3 border-t border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 space-y-2">
                  <div className="flex gap-2">
                    <Textarea
                      placeholder="Nhập nội dung phản hồi..."
                      value={replyMessage}
                      onChange={(e) => setReplyMessage(e.target.value)}
                      rows={2}
                      className="text-xs"
                    />
                    <Button
                      size="sm"
                      loading={replyMutation.isPending}
                      disabled={!replyMessage.trim()}
                      onClick={() => replyMutation.mutate()}
                      className="self-end"
                    >
                      <Send className="h-4 w-4" />
                    </Button>
                  </div>

                  {/* Drag and drop / attachment picker */}
                  <div className="flex items-center justify-between text-xs text-slate-500">
                    <label className="inline-flex items-center gap-1.5 cursor-pointer text-teal-600 hover:text-teal-700 dark:text-teal-400">
                      <Paperclip className="h-3.5 w-3.5" />
                      <span>Đính kèm tệp tin (ảnh, PDF &lt; 10MB)</span>
                      <input
                        type="file"
                        className="hidden"
                        onChange={(e) => {
                          const file = e.target.files?.[0];
                          if (file) handleFileUpload(file);
                        }}
                      />
                    </label>

                    {uploading && (
                      <span className="text-teal-600 animate-pulse text-[11px]">
                        Đang tải tệp lên...
                      </span>
                    )}
                  </div>
                </div>
              ) : (
                <div className="p-3 bg-slate-100 text-center text-xs text-slate-500 dark:bg-slate-800">
                  Yêu cầu này đã được đóng. Bạn không thể gửi thêm phản hồi.
                </div>
              )}
            </div>
          ) : (
            <div className="rounded-card border border-dashed border-slate-300 p-16 text-center text-xs text-slate-500 dark:border-slate-800">
              <Headphones className="mx-auto h-10 w-10 text-slate-400 mb-2 opacity-50" />
              Chọn một yêu cầu ở danh sách bên trái hoặc bấm "Tạo yêu cầu mới" để bắt đầu hội thoại.
            </div>
          )}
        </div>
      </div>

      {/* New Ticket Modal */}
      <Dialog
        open={createModalOpen}
        onOpenChange={setCreateModalOpen}
        title="Tạo Yêu cầu Hỗ trợ Mới"
        description="Điền tiêu đề và mô tả chi tiết vấn đề bạn đang gặp phải với bài lab hoặc đơn hàng."
      >
        <form
          onSubmit={(e) => {
            e.preventDefault();
            createMutation.mutate();
          }}
          className="space-y-4"
        >
          {createError && (
            <div className="rounded bg-rose-50 p-2.5 text-xs text-rose-700 dark:bg-rose-950/50 dark:text-rose-300 flex items-center gap-1.5">
              <AlertCircle className="h-4 w-4 shrink-0 text-rose-500" />
              <span>{createError}</span>
            </div>
          )}

          <div className="space-y-1.5">
            <label className="text-xs font-semibold text-slate-700 dark:text-slate-300">
              Chủ đề yêu cầu *
            </label>
            <Input
              required
              placeholder="vd: Cần giải đáp kịch bản C2 Cache Rule"
              value={newSubject}
              onChange={(e) => setNewSubject(e.target.value)}
              className="text-xs"
            />
          </div>

          <div className="space-y-1.5">
            <label className="text-xs font-semibold text-slate-700 dark:text-slate-300">
              Nội dung chi tiết *
            </label>
            <Textarea
              required
              placeholder="Mô tả cụ thể thông báo lỗi hoặc các bước bạn đã thực hiện..."
              value={newInitialMsg}
              onChange={(e) => setNewInitialMsg(e.target.value)}
              rows={4}
              className="text-xs"
            />
          </div>

          <div className="flex justify-end gap-2 pt-2 border-t border-slate-100 dark:border-slate-800">
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={() => setCreateModalOpen(false)}
            >
              Hủy
            </Button>
            <Button
              type="submit"
              size="sm"
              loading={createMutation.isPending}
              disabled={!newSubject.trim() || !newInitialMsg.trim()}
            >
              Gửi yêu cầu
            </Button>
          </div>
        </form>
      </Dialog>
    </div>
  );
}
