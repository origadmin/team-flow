# 详细项目优化计划

## 1. 项目现状评估

### 1.1 核心问题
- **社交粘性缺失**：订阅/关注体系不完整，互动反馈闭环碎片化，推荐与发现功能薄弱
- **媒体处理深度不足**：视频预览精灵图缺失，多语言与字幕支持不足，媒体元数据挖掘表面化，视频章节功能缺失
- **CMS管理能力薄弱**：审核流缺失，仪表盘功能不足，分类体系不完善
- **工程化与产品化断层**：过度关注框架，前后端割裂严重，Mock数据泛滥

### 1.2 目标状态
- 完整的社交体系：用户订阅、互动反馈、推荐系统
- 专业的媒体处理：预览精灵图、多语言字幕、元数据挖掘、视频章节
- 强大的CMS管理能力：审核流、数据仪表盘、完善的分类体系
- 工程化与产品化：消除Mock数据，统一API规范，高质量代码结构

## 2. 详细优化计划

### 2.1 第一阶段：基础优化（1周）

#### 2.1.1 任务1：集中管理配置

**技术设计**：
- 创建 `src/config` 目录，包含 `index.ts` 和 `types.ts` 文件
- 使用环境变量和默认值相结合的方式管理配置
- 实现配置的类型定义，确保类型安全

**具体步骤**：
1. 创建 `src/config/types.ts` 文件，定义配置类型
2. 创建 `src/config/index.ts` 文件，实现配置管理逻辑
3. 替换所有文件中的硬编码配置
4. 测试配置加载和环境变量支持

**代码结构**：
```typescript
// src/config/types.ts
export interface Config {
  api: {
    baseUrl: string;
    prefix: string;
    timeout: number;
  };
  app: {
    name: string;
    version: string;
  };
}

// src/config/index.ts
import type { Config } from './types';

const config: Config = {
  api: {
    baseUrl: import.meta.env.VITE_API_BASE_URL || 'http://localhost:9090',
    prefix: '/api/v1',
    timeout: 30000,
  },
  app: {
    name: 'OrigCMS',
    version: '1.0.0',
  },
};

export default config;
```

**验收标准**：
- 所有文件使用统一的配置管理
- 配置能够正确加载环境变量
- 配置类型定义完整

#### 2.1.2 任务2：消除重复代码

**技术设计**：
- 提取 `getFullUrl` 函数到 `src/lib/utils.ts`
- 统一API调用模式，创建基础API服务
- 优化组件结构，提取共用组件

**具体步骤**：
1. 提取 `getFullUrl` 函数到 `src/lib/utils.ts`
2. 创建 `src/lib/api/base.ts` 文件，统一API调用模式
3. 识别并提取共用组件到 `src/components/common` 目录
4. 替换所有文件中的重复代码

**代码结构**：
```typescript
// src/lib/utils.ts
export const getFullUrl = (path?: string): string => {
  if (!path) return '';
  if (path.startsWith('http')) return path;
  const config = require('../config').default;
  const base = config.api.baseUrl.replace(/\/$/, '');
  const sep = path.startsWith('/') ? '' : '/';
  return `${base}${sep}${path}`;
};

// src/lib/api/base.ts
import config from '../config';
import axios from 'axios';

const apiClient = axios.create({
  baseURL: config.api.baseUrl + config.api.prefix,
  timeout: config.api.timeout,
  headers: {
    'Content-Type': 'application/json',
  },
});

export default apiClient;
```

**验收标准**：
- 消除所有重复代码
- API调用模式统一
- 共用组件提取完成

#### 2.1.3 任务3：修复兼容性问题

**技术设计**：
- 分析React Router与React 19的兼容性问题
- 实现更稳定的路由参数获取方法
- 确保所有页面能够正常加载

**具体步骤**：
1. 分析ProfilePage中的路由参数获取问题
2. 实现基于useLocation的路由参数获取方法
3. 测试所有页面的路由功能
4. 修复可能的路由兼容性问题

**代码结构**：
```typescript
// src/pages/home/Profile.tsx
import { useLocation } from '@tanstack/react-router';

const ProfilePage = () => {
  let id = '1';
  try {
    const location = useLocation();
    const pathParts = location.pathname.split('/');
    if (pathParts.length > 2 && pathParts[1] === 'u') {
      id = pathParts[2] || '1';
    }
  } catch (error) {
    console.error('Error getting location:', error);
  }
  // 其他代码
};
```

**验收标准**：
- 所有页面能够正常加载，无路由错误
- 路由参数能够正确获取
- 路由切换功能正常

#### 2.1.4 任务4：消除Mock数据

**技术设计**：
- 识别所有使用Mock数据的页面
- 对接真实API
- 添加错误处理和加载状态

**具体步骤**：
1. 识别所有使用Mock数据的页面
2. 实现真实API调用
3. 添加错误处理和加载状态
4. 测试API调用的正确性

**代码结构**：
```typescript
// src/pages/home/Members.tsx
import { useState, useEffect } from 'react';
import { userApi } from '@/lib/api/user';

const MembersPage = () => {
  const [members, setMembers] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchMembers = async () => {
      try {
        setLoading(true);
        const response = await userApi.list({ page_size: 100 });
        setMembers(response.list || []);
      } catch (err) {
        setError('Failed to fetch members');
        console.error('Failed to fetch members:', err);
      } finally {
        setLoading(false);
      }
    };

    fetchMembers();
  }, []);

  // 其他代码
};
```

**验收标准**：
- 所有页面使用真实API数据
- API调用有错误处理和加载状态
- 页面能够正常显示数据

### 2.2 第二阶段：社交体系构建（2周）

#### 2.2.1 任务1：用户订阅功能

**技术设计**：
- 后端实现订阅/关注API
- 前端实现订阅按钮和状态管理
- 实现订阅列表和通知

**具体步骤**：
1. 后端实现订阅/关注API
2. 前端实现订阅按钮和状态管理
3. 实现订阅列表页面
4. 测试订阅功能

**代码结构**：
```typescript
// 后端API
// internal/server/subscription.go
func (s *Server) handleSubscribe(w http.ResponseWriter, r *http.Request) {
  // 实现订阅逻辑
}

// 前端组件
// src/components/common/SubscribeButton.tsx
import { useState, useEffect } from 'react';
import { subscriptionApi } from '@/lib/api/subscription';

const SubscribeButton = ({ userId }: { userId: string }) => {
  const [isSubscribed, setIsSubscribed] = useState(false);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    const checkSubscription = async () => {
      try {
        const status = await subscriptionApi.getStatus(userId);
        setIsSubscribed(status.is_subscribed);
      } catch (err) {
        console.error('Failed to check subscription status:', err);
      }
    };

    checkSubscription();
  }, [userId]);

  const handleSubscribe = async () => {
    try {
      setLoading(true);
      if (isSubscribed) {
        await subscriptionApi.unsubscribe(userId);
      } else {
        await subscriptionApi.subscribe(userId);
      }
      setIsSubscribed(!isSubscribed);
    } catch (err) {
      console.error('Failed to toggle subscription:', err);
    } finally {
      setLoading(false);
    }
  };

  return (
    <button
      onClick={handleSubscribe}
      disabled={loading}
      className="px-4 py-2 bg-blue-600 text-white rounded-full hover:bg-blue-700"
    >
      {loading ? 'Loading...' : isSubscribed ? 'Subscribed' : 'Subscribe'}
    </button>
  );
};
```

**验收标准**：
- 用户能够订阅/取消订阅其他用户
- 订阅状态能够正确显示
- 订阅列表能够正常显示

#### 2.2.2 任务2：互动反馈系统

**技术设计**：
- 后端实现评论、点赞、分享API
- 前端实现评论区和互动按钮
- 实现评论回复和点赞计数

**具体步骤**：
1. 后端实现评论、点赞、分享API
2. 前端实现评论区组件
3. 实现互动按钮（点赞、分享）
4. 测试互动功能

**代码结构**：
```typescript
// 后端API
// internal/server/comment.go
func (s *Server) handleCreateComment(w http.ResponseWriter, r *http.Request) {
  // 实现评论创建逻辑
}

// 前端组件
// src/components/common/CommentSection.tsx
import { useState, useEffect } from 'react';
import { commentApi } from '@/lib/api/comment';

const CommentSection = ({ mediaId }: { mediaId: string }) => {
  const [comments, setComments] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [commentText, setCommentText] = useState('');

  useEffect(() => {
    const fetchComments = async () => {
      try {
        setLoading(true);
        const response = await commentApi.getAll({ media_id: mediaId });
        setComments(response.list || []);
      } catch (err) {
        setError('Failed to fetch comments');
        console.error('Failed to fetch comments:', err);
      } finally {
        setLoading(false);
      }
    };

    fetchComments();
  }, [mediaId]);

  const handleSubmitComment = async () => {
    if (!commentText.trim()) return;

    try {
      await commentApi.create({
        media_id: mediaId,
        body: commentText,
      });
      setCommentText('');
      // 重新获取评论
      const response = await commentApi.getAll({ media_id: mediaId });
      setComments(response.list || []);
    } catch (err) {
      console.error('Failed to submit comment:', err);
    }
  };

  return (
    <div className="space-y-4">
      <div className="flex">
        <input
          type="text"
          value={commentText}
          onChange={(e) => setCommentText(e.target.value)}
          placeholder="Add a comment..."
          className="flex-1 px-4 py-2 border rounded-full"
        />
        <button
          onClick={handleSubmitComment}
          className="ml-2 px-4 py-2 bg-blue-600 text-white rounded-full"
        >
          Post
        </button>
      </div>
      {loading ? (
        <div>Loading comments...</div>
      ) : error ? (
        <div className="text-red-500">{error}</div>
      ) : (
        <div className="space-y-4">
          {comments.map((comment) => (
            <div key={comment.id} className="p-4 border rounded-lg">
              <div className="font-bold">{comment.username}</div>
              <div>{comment.body}</div>
              <div className="flex items-center mt-2">
                <button className="mr-4 text-gray-500">Like</button>
                <button className="text-gray-500">Reply</button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};
```

**验收标准**：
- 用户能够评论、点赞、分享视频
- 评论回复功能正常
- 点赞计数能够正确显示

#### 2.2.3 任务3：通知系统

**技术设计**：
- 后端实现通知API
- 前端实现通知中心
- 实现实时通知推送

**具体步骤**：
1. 后端实现通知API
2. 前端实现通知中心组件
3. 实现实时通知推送
4. 测试通知功能

**代码结构**：
```typescript
// 后端API
// internal/server/notification.go
func (s *Server) handleGetNotifications(w http.ResponseWriter, r *http.Request) {
  // 实现获取通知逻辑
}

// 前端组件
// src/components/portal/NotificationCenter.tsx
import { useState, useEffect } from 'react';
import { notificationApi } from '@/lib/api/notification';

const NotificationCenter = () => {
  const [notifications, setNotifications] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchNotifications = async () => {
      try {
        setLoading(true);
        const response = await notificationApi.getAll();
        setNotifications(response.list || []);
      } catch (err) {
        setError('Failed to fetch notifications');
        console.error('Failed to fetch notifications:', err);
      } finally {
        setLoading(false);
      }
    };

    fetchNotifications();
  }, []);

  return (
    <div className="space-y-4">
      <h3 className="font-bold text-lg">Notifications</h3>
      {loading ? (
        <div>Loading notifications...</div>
      ) : error ? (
        <div className="text-red-500">{error}</div>
      ) : notifications.length === 0 ? (
        <div>No notifications</div>
      ) : (
        <div className="space-y-4">
          {notifications.map((notification) => (
            <div key={notification.id} className="p-4 border rounded-lg">
              <div className="font-bold">{notification.title}</div>
              <div>{notification.message}</div>
              <div className="text-sm text-gray-500">{notification.created_at}</div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};
```

**验收标准**：
- 用户能够收到订阅、评论、点赞等通知
- 通知中心能够正常显示
- 实时通知推送功能正常

### 2.3 第三阶段：媒体处理专业化（2周）

#### 2.3.1 任务1：视频预览精灵图

**技术设计**：
- 后端实现视频帧提取和精灵图生成
- 前端实现鼠标悬停预览功能
- 测试不同视频格式的兼容性

**具体步骤**：
1. 后端实现视频帧提取和精灵图生成
2. 前端实现鼠标悬停预览功能
3. 测试不同视频格式的兼容性
4. 优化精灵图生成和预览性能

**代码结构**：
```go
// 后端代码
// internal/svc-media/biz/sprite.go
func (b *MediaBiz) GenerateSprite(mediaID int64) error {
  // 实现视频帧提取和精灵图生成逻辑
}

// 前端组件
// src/components/common/VideoPreview.tsx
import { useState } from 'react';

const VideoPreview = ({ spriteUrl, duration }: { spriteUrl: string; duration: number }) => {
  const [hoverTime, setHoverTime] = useState<number | null>(null);

  const handleMouseMove = (e: React.MouseEvent<HTMLDivElement>) => {
    const rect = e.currentTarget.getBoundingClientRect();
    const position = (e.clientX - rect.left) / rect.width;
    const time = position * duration;
    setHoverTime(time);
  };

  const handleMouseLeave = () => {
    setHoverTime(null);
  };

  return (
    <div
      className="relative"
      onMouseMove={handleMouseMove}
      onMouseLeave={handleMouseLeave}
    >
      <div className="w-full h-1 bg-gray-300 rounded-full">
        <div className="h-full bg-blue-600 rounded-full" style={{ width: '50%' }} />
      </div>
      {hoverTime !== null && (
        <div className="absolute bottom-4 left-0 right-0 flex justify-center">
          <div className="bg-black/80 text-white p-2 rounded">
            <img
              src={spriteUrl}
              alt="Preview"
              className="w-32 h-18 object-cover"
            />
          </div>
        </div>
      )}
    </div>
  );
};
```

**验收标准**：
- 鼠标悬停在视频进度条上时显示预览图
- 不同视频格式的兼容性良好
- 精灵图生成和预览性能良好

#### 2.3.2 任务2：多语言字幕支持

**技术设计**：
- 后端实现字幕上传和管理API
- 前端实现字幕选择和显示
- 集成Whisper API实现自动转码

**具体步骤**：
1. 后端实现字幕上传和管理API
2. 前端实现字幕选择和显示组件
3. 集成Whisper API实现自动转码
4. 测试字幕功能

**代码结构**：
```typescript
// 后端API
// internal/server/subtitle.go
func (s *Server) handleUploadSubtitle(w http.ResponseWriter, r *http.Request) {
  // 实现字幕上传逻辑
}

// 前端组件
// src/components/common/SubtitleSelector.tsx
import { useState, useEffect } from 'react';
import { subtitleApi } from '@/lib/api/subtitle';

const SubtitleSelector = ({ mediaId }: { mediaId: string }) => {
  const [subtitles, setSubtitles] = useState<any[]>([]);
  const [selectedSubtitle, setSelectedSubtitle] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchSubtitles = async () => {
      try {
        setLoading(true);
        const response = await subtitleApi.getByMediaId(mediaId);
        setSubtitles(response.list || []);
      } catch (err) {
        console.error('Failed to fetch subtitles:', err);
      } finally {
        setLoading(false);
      }
    };

    fetchSubtitles();
  }, [mediaId]);

  const handleSubtitleChange = (subtitleId: string) => {
    setSelectedSubtitle(subtitleId);
  };

  return (
    <div className="space-y-2">
      <label className="font-medium">Subtitles:</label>
      {loading ? (
        <div>Loading subtitles...</div>
      ) : subtitles.length === 0 ? (
        <div>No subtitles available</div>
      ) : (
        <select
          value={selectedSubtitle || ''}
          onChange={(e) => handleSubtitleChange(e.target.value)}
          className="px-4 py-2 border rounded"
        >
          <option value="">None</option>
          {subtitles.map((subtitle) => (
            <option key={subtitle.id} value={subtitle.id}>
              {subtitle.language} - {subtitle.name}
            </option>
          ))}
        </select>
      )}
    </div>
  );
};
```

**验收标准**：
- 用户能够上传、选择和显示多语言字幕
- 自动字幕转码功能正常
- 字幕显示效果良好

#### 2.3.3 任务3：媒体元数据挖掘

**技术设计**：
- 后端实现元数据提取API
- 前端实现元数据展示
- 测试不同媒体类型的元数据提取

**具体步骤**：
1. 后端实现元数据提取API
2. 前端实现元数据展示组件
3. 测试不同媒体类型的元数据提取
4. 优化元数据提取性能

**代码结构**：
```go
// 后端代码
// internal/svc-media/biz/metadata.go
func (b *MediaBiz) ExtractMetadata(mediaID int64) error {
  // 实现元数据提取逻辑
}

// 前端组件
// src/components/common/MediaMetadata.tsx
import { useState, useEffect } from 'react';
import { mediaApi } from '@/lib/api/media';

const MediaMetadata = ({ mediaId }: { mediaId: string }) => {
  const [metadata, setMetadata] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchMetadata = async () => {
      try {
        setLoading(true);
        const response = await mediaApi.getMetadata(mediaId);
        setMetadata(response);
      } catch (err) {
        setError('Failed to fetch metadata');
        console.error('Failed to fetch metadata:', err);
      } finally {
        setLoading(false);
      }
    };

    fetchMetadata();
  }, [mediaId]);

  return (
    <div className="space-y-4">
      <h3 className="font-bold text-lg">Metadata</h3>
      {loading ? (
        <div>Loading metadata...</div>
      ) : error ? (
        <div className="text-red-500">{error}</div>
      ) : metadata ? (
        <div className="space-y-2">
          {metadata.duration && <div><strong>Duration:</strong> {metadata.duration} seconds</div>}
          {metadata.resolution && <div><strong>Resolution:</strong> {metadata.resolution}</div>}
          {metadata.format && <div><strong>Format:</strong> {metadata.format}</div>}
          {metadata.codec && <div><strong>Codec:</strong> {metadata.codec}</div>}
          {metadata.bitrate && <div><strong>Bitrate:</strong> {metadata.bitrate} kbps</div>}
        </div>
      ) : (
        <div>No metadata available</div>
      )}
    </div>
  );
};
```

**验收标准**：
- 系统能够提取和展示图片EXIF、音频ID3、视频分辨率等元数据
- 不同媒体类型的元数据提取效果良好
- 元数据提取性能良好

### 2.4 第四阶段：CMS管理能力（1周）

#### 2.4.1 任务1：审核流

**技术设计**：
- 后端实现审核API和状态管理
- 前端实现审核界面
- 测试审核流程

**具体步骤**：
1. 后端实现审核API和状态管理
2. 前端实现审核界面
3. 测试审核流程
4. 优化审核体验

**代码结构**：
```typescript
// 后端API
// internal/server/moderation.go
func (s *Server) handleApproveMedia(w http.ResponseWriter, r *http.Request) {
  // 实现审核通过逻辑
}

// 前端组件
// src/pages/admin/MediaModeration.tsx
import { useState, useEffect } from 'react';
import { mediaApi } from '@/lib/api/media';

const MediaModeration = () => {
  const [mediaList, setMediaList] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchPendingMedia = async () => {
      try {
        setLoading(true);
        const response = await mediaApi.list({ status: 'pending' });
        setMediaList(response.list || []);
      } catch (err) {
        setError('Failed to fetch pending media');
        console.error('Failed to fetch pending media:', err);
      } finally {
        setLoading(false);
      }
    };

    fetchPendingMedia();
  }, []);

  const handleApprove = async (mediaId: string) => {
    try {
      await mediaApi.approve(mediaId);
      // 重新获取待审核媒体
      const response = await mediaApi.list({ status: 'pending' });
      setMediaList(response.list || []);
    } catch (err) {
      console.error('Failed to approve media:', err);
    }
  };

  const handleReject = async (mediaId: string) => {
    try {
      await mediaApi.reject(mediaId);
      // 重新获取待审核媒体
      const response = await mediaApi.list({ status: 'pending' });
      setMediaList(response.list || []);
    } catch (err) {
      console.error('Failed to reject media:', err);
    }
  };

  return (
    <div className="space-y-4">
      <h3 className="font-bold text-lg">Pending Media</h3>
      {loading ? (
        <div>Loading pending media...</div>
      ) : error ? (
        <div className="text-red-500">{error}</div>
      ) : mediaList.length === 0 ? (
        <div>No pending media</div>
      ) : (
        <div className="space-y-4">
          {mediaList.map((media) => (
            <div key={media.id} className="p-4 border rounded-lg">
              <div className="font-bold">{media.title}</div>
              <div className="text-sm text-gray-500">{media.username}</div>
              <div className="mt-2">
                <button
                  onClick={() => handleApprove(media.id)}
                  className="mr-2 px-4 py-2 bg-green-600 text-white rounded"
                >
                  Approve
                </button>
                <button
                  onClick={() => handleReject(media.id)}
                  className="px-4 py-2 bg-red-600 text-white rounded"
                >
                  Reject
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};
```

**验收标准**：
- 管理员能够审核、批准或拒绝上传的内容
- 审核流程能够正常运行
- 审核状态能够正确更新

#### 2.4.2 任务2：数据仪表盘

**技术设计**：
- 后端实现统计API
- 前端实现仪表盘界面
- 测试数据准确性

**具体步骤**：
1. 后端实现统计API
2. 前端实现仪表盘界面
3. 测试数据准确性
4. 优化仪表盘性能

**代码结构**：
```typescript
// 后端API
// internal/server/stats.go
func (s *Server) handleGetStats(w http.ResponseWriter, r *http.Request) {
  // 实现统计逻辑
}

// 前端组件
// src/pages/admin/Dashboard.tsx
import { useState, useEffect } from 'react';
import { statsApi } from '@/lib/api/stats';

const Dashboard = () => {
  const [stats, setStats] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchStats = async () => {
      try {
        setLoading(true);
        const response = await statsApi.getDashboard();
        setStats(response);
      } catch (err) {
        setError('Failed to fetch stats');
        console.error('Failed to fetch stats:', err);
      } finally {
        setLoading(false);
      }
    };

    fetchStats();
  }, []);

  return (
    <div className="space-y-6">
      <h2 className="text-2xl font-bold">Dashboard</h2>
      {loading ? (
        <div>Loading stats...</div>
      ) : error ? (
        <div className="text-red-500">{error}</div>
      ) : stats ? (
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <div className="p-4 bg-white border rounded-lg shadow">
            <div className="text-sm text-gray-500">Total Users</div>
            <div className="text-2xl font-bold">{stats.total_users}</div>
          </div>
          <div className="p-4 bg-white border rounded-lg shadow">
            <div className="text-sm text-gray-500">Total Media</div>
            <div className="text-2xl font-bold">{stats.total_media}</div>
          </div>
          <div className="p-4 bg-white border rounded-lg shadow">
            <div className="text-sm text-gray-500">Total Views</div>
            <div className="text-2xl font-bold">{stats.total_views}</div>
          </div>
          <div className="p-4 bg-white border rounded-lg shadow">
            <div className="text-sm text-gray-500">Pending Media</div>
            <div className="text-2xl font-bold">{stats.pending_media}</div>
          </div>
        </div>
      ) : (
        <div>No stats available</div>
      )}
    </div>
  );
};
```

**验收标准**：
- 管理员能够查看流量统计、存储消耗、用户增长等数据
- 数据仪表盘能够正常显示
- 数据准确性良好

#### 2.4.3 任务3：分类体系

**技术设计**：
- 后端实现分类和标签API
- 前端实现分类管理界面
- 测试分类和标签的联动

**具体步骤**：
1. 后端实现分类和标签API
2. 前端实现分类管理界面
3. 测试分类和标签的联动
4. 优化分类管理体验

**代码结构**：
```typescript
// 后端API
// internal/server/category.go
func (s *Server) handleCreateCategory(w http.ResponseWriter, r *http.Request) {
  // 实现分类创建逻辑
}

// 前端组件
// src/pages/admin/Categories.tsx
import { useState, useEffect } from 'react';
import { categoryApi } from '@/lib/api/category';

const Categories = () => {
  const [categories, setCategories] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [newCategory, setNewCategory] = useState('');

  useEffect(() => {
    const fetchCategories = async () => {
      try {
        setLoading(true);
        const response = await categoryApi.list();
        setCategories(response.list || []);
      } catch (err) {
        setError('Failed to fetch categories');
        console.error('Failed to fetch categories:', err);
      } finally {
        setLoading(false);
      }
    };

    fetchCategories();
  }, []);

  const handleCreateCategory = async () => {
    if (!newCategory.trim()) return;

    try {
      await categoryApi.create({ name: newCategory });
      setNewCategory('');
      // 重新获取分类
      const response = await categoryApi.list();
      setCategories(response.list || []);
    } catch (err) {
      console.error('Failed to create category:', err);
    }
  };

  return (
    <div className="space-y-4">
      <h3 className="font-bold text-lg">Categories</h3>
      <div className="flex">
        <input
          type="text"
          value={newCategory}
          onChange={(e) => setNewCategory(e.target.value)}
          placeholder="Add a category..."
          className="flex-1 px-4 py-2 border rounded-l"
        />
        <button
          onClick={handleCreateCategory}
          className="px-4 py-2 bg-blue-600 text-white rounded-r"
        >
          Add
        </button>
      </div>
      {loading ? (
        <div>Loading categories...</div>
      ) : error ? (
        <div className="text-red-500">{error}</div>
      ) : categories.length === 0 ? (
        <div>No categories</div>
      ) : (
        <div className="space-y-2">
          {categories.map((category) => (
            <div key={category.id} className="flex items-center justify-between p-4 border rounded">
              <div>{category.name}</div>
              <div className="flex">
                <button className="mr-2 px-2 py-1 bg-blue-600 text-white rounded">Edit</button>
                <button className="px-2 py-1 bg-red-600 text-white rounded">Delete</button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};
```

**验收标准**：
- 管理员能够创建、编辑、删除分类和标签
- 内容能够正确分类
- 分类和标签的联动功能正常

### 2.5 第五阶段：测试与优化（1周）

#### 2.5.1 任务1：测试

**技术设计**：
- 实现单元测试
- 添加集成测试
- 进行性能测试

**具体步骤**：
1. 实现单元测试
2. 添加集成测试
3. 进行性能测试
4. 修复测试中发现的问题

**代码结构**：
```typescript
// 单元测试
// src/lib/utils.test.ts
import { getFullUrl } from './utils';

describe('getFullUrl', () => {
  it('should return empty string for empty path', () => {
    expect(getFullUrl('')).toBe('');
  });

  it('should return the same URL for absolute URLs', () => {
    const url = 'https://example.com/path';
    expect(getFullUrl(url)).toBe(url);
  });

  it('should prepend base URL for relative paths', () => {
    const baseUrl = 'http://localhost:9090';
    const path = '/api/v1/media';
    expect(getFullUrl(path)).toBe(`${baseUrl}${path}`);
  });
});

// 集成测试
// src/pages/home/Watch.test.tsx
import { render, screen } from '@testing-library/react';
import WatchPage from './Watch';

jest.mock('@/hooks/queries', () => ({
  useMediaDetail: () => ({
    data: {
      id: '1',
      title: 'Test Video',
      description: 'Test Description',
      view_count: 100,
      created_at: '2024-01-01T00:00:00Z',
    },
    isLoading: false,
    error: null,
  }),
}));

describe('WatchPage', () => {
  it('should render video title', () => {
    render(<WatchPage />);
    expect(screen.getByText('Test Video')).toBeInTheDocument();
  });

  it('should render video description', () => {
    render(<WatchPage />);
    expect(screen.getByText('Test Description')).toBeInTheDocument();
  });
});
```

**验收标准**：
- 测试覆盖率达到80%以上
- 无严重bug
- 性能测试通过

#### 2.5.2 任务2：性能优化

**技术设计**：
- 优化数据库查询
- 实现缓存策略
- 优化前端渲染

**具体步骤**：
1. 优化数据库查询
2. 实现缓存策略
3. 优化前端渲染
4. 测试性能优化效果

**代码结构**：
```go
// 后端代码
// internal/svc-media/data/media_repo.go
func (r *MediaRepo) ListMediaWithCache(params MediaListParams) ([]*Media, error) {
  // 实现缓存策略
}

// 前端代码
// src/components/common/VideoCard.tsx
import { memo } from 'react';

const VideoCard = memo(({ video }: { video: any }) => {
  // 组件代码
});

export default VideoCard;
```

**验收标准**：
- 系统响应时间减少30%以上
- 页面加载时间 < 2秒
- API响应时间 < 500ms

#### 2.5.3 任务3：安全增强

**技术设计**：
- 实现输入验证
- 增强Token管理
- 添加防XSS措施

**具体步骤**：
1. 实现输入验证
2. 增强Token管理
3. 添加防XSS措施
4. 进行安全扫描

**代码结构**：
```typescript
// 前端代码
// src/lib/validators.ts
export const validateEmail = (email: string): boolean => {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  return emailRegex.test(email);
};

export const validatePassword = (password: string): boolean => {
  return password.length >= 8;
};

// 后端代码
// internal/middleware/security.go
func SecurityMiddleware(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    // 实现安全措施
    next.ServeHTTP(w, r);
  });
}
```

**验收标准**：
- 通过OWASP ZAP安全扫描
- 无严重安全漏洞
- 数据传输加密

## 3. 资源需求

### 3.1 人员
- 前端开发：2人
- 后端开发：2人
- 测试团队：1人
- 安全团队：1人

### 3.2 技术资源
- 服务器：至少2台（开发和测试）
- 存储：足够的存储空间用于媒体文件
- 第三方服务：Whisper API（用于自动字幕转码）

### 3.3 工具
- 版本控制：Git
- 项目管理：Jira
- 测试工具：Jest、Cypress
- 性能分析：Lighthouse
- 安全扫描：OWASP ZAP

## 4. 风险评估

### 4.1 潜在风险
1. **技术风险**：视频处理复杂度高，可能需要专业知识
2. **时间风险**：社交体系和媒体处理功能可能需要更多时间
3. **资源风险**：服务器和存储资源可能不足
4. **集成风险**：前后端集成可能遇到接口不一致问题

### 4.2 风险应对措施
1. **技术风险**：寻求专业媒体处理知识支持，使用成熟的开源库
2. **时间风险**：分阶段实施，优先实现核心功能，设置合理的时间缓冲
3. **资源风险**：提前评估存储需求，使用云存储服务
4. **集成风险**：建立严格的API规范，定期进行接口测试

## 5. 里程碑

| 阶段 | 完成标志 | 时间 |
|------|----------|------|
| 基础优化 | 消除Mock数据，修复兼容性问题 | 第1周末 |
| 社交体系构建 | 实现订阅、评论、通知功能 | 第3周末 |
| 媒体处理专业化 | 实现预览精灵图、字幕支持、元数据挖掘 | 第5周末 |
| CMS管理能力 | 实现审核流、仪表盘、分类体系 | 第6周末 |
| 测试与优化 | 通过测试，性能和安全达标 | 第7周末 |

## 6. 验收标准

### 6.1 功能验收
- 社交体系：用户能够订阅、评论、点赞、分享，收到通知
- 媒体处理：视频有预览精灵图，支持多语言字幕，显示元数据
- CMS管理：管理员能够审核内容，查看数据仪表盘，管理分类
- 工程化：所有页面使用真实API数据，代码质量高

### 6.2 性能验收
- 页面加载时间：首屏加载时间 < 2秒
- 视频播放：流畅无卡顿
- API响应：平均响应时间 < 500ms

### 6.3 安全验收
- 通过OWASP ZAP安全扫描
- 无严重安全漏洞
- 数据传输加密

## 7. 结论

本计划提供了一个详细的、可执行的项目优化方案，从基础优化到功能完善，再到测试与优化，覆盖了项目的各个方面。通过实施本计划，项目将从一个技术原型转变为一个功能完整、体验优秀的MediaCMS产品，真正实现"以内容为中心、以用户互动为纽带"的视频内容社区。

计划的成功实施需要团队成员的密切配合和持续努力，同时需要根据实际情况进行适当的调整。通过设定明确的目标和验收标准，确保项目优化工作能够按计划完成，达到预期的成果。