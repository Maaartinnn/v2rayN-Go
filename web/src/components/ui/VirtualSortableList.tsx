import React, { useState, useCallback, useRef, useMemo } from 'react'
import { useVirtualizer } from '@tanstack/react-virtual'
import {
  DndContext,
  closestCenter,
  KeyboardSensor,
  PointerSensor,
  useSensor,
  useSensors,
  DragOverlay,
  defaultDropAnimationSideEffects,
  type DragEndEvent,
  type DragStartEvent,
  type UniqueIdentifier,
} from '@dnd-kit/core'
import {
  arrayMove,
  SortableContext,
  sortableKeyboardCoordinates,
  useSortable,
  verticalListSortingStrategy,
} from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'
import { GripVertical } from 'lucide-react'

// ==========================================
// 类型定义
// ==========================================

export interface VirtualSortableListProps<T extends { uuid: string }> {
  /** 数据列表 */
  items: T[]
  /** 列表更新回调（拖拽排序后） */
  onItemsChange: (newItems: T[]) => void
  /** 渲染单个卡片 */
  renderItem: (props: {
    item: T
    isDragging: boolean
    isOverlay: boolean
    dragListeners: Record<string, any>
    dragAttributes: Record<string, any>
  }) => React.ReactNode
  /** 预估行高（像素），默认 72 */
  estimateSize?: number
  /** 预渲染数量，默认 5 */
  overscan?: number
  /** 外层容器类名 */
  className?: string
  /** 外层容器样式 */
  style?: React.CSSProperties
  /** 空状态内容 */
  emptyContent?: React.ReactNode
  /** 拖拽开始时的回调（用于关闭编辑/删除状态） */
  onDragStart?: () => void
  /** 是否禁用拖拽 */
  disableDrag?: boolean
  /** 额外渲染在卡片下方的内容（展开面板等） */
  renderExtra?: (item: T) => React.ReactNode
  /** 判断某个 item 的额外内容是否处于展开/激活状态（用于提升 zIndex） */
  isItemExpanded?: (item: T) => boolean
}

// ==========================================
// SortableVirtualItem: 双层容器（虚拟层 + 拖拽层）
// ==========================================

interface SortableVirtualItemProps<T extends { uuid: string }> {
  item: T
  virtualItem: { index: number; start: number; end: number; size: number; key: React.Key }
  virtualizer: { measureElement: (element: Element | null) => void; getTotalSize: () => number }
  renderItem: VirtualSortableListProps<T>['renderItem']
  renderExtra?: VirtualSortableListProps<T>['renderExtra']
  disableDrag?: boolean
  isExpanded?: boolean
}

function SortableVirtualItem<T extends { uuid: string }>({
  item,
  virtualItem,
  virtualizer,
  renderItem,
  renderExtra,
  disableDrag = false,
  isExpanded = false,
}: SortableVirtualItemProps<T>) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({
    id: item.uuid,
    disabled: disableDrag,
  })

  const extraContent = renderExtra?.(item)

  // 【修复9】：外层（虚拟层）只负责测量和定位，不做 transform
  // 内层（DND 层）负责 transform 变形，不影响 measureElement 的测量结果
  return (
    // 外层容器 - Virtual Layer: 负责高度测算和绝对定位
    <div
      ref={virtualizer.measureElement}
      data-index={virtualItem.index}
      style={{
        position: 'absolute',
        top: `${virtualItem.start}px`,
        left: 0,
        width: '100%',
        // 【修复3】：支持 isExpanded 提升 zIndex，避免直接 DOM 操作被 React 重渲染覆盖
        zIndex: isDragging ? 50 : isExpanded ? 40 : 1,
      }}
    >
      {/* 内层容器 - DND Layer: 负责拖拽时的物理变形和占位 */}
      <div
        ref={setNodeRef}
        style={{
          transform: CSS.Translate.toString(transform),
          transition,
          opacity: isDragging ? 0.3 : 1, // 让原位置的占位符半透明
          marginBottom: '8px',
        }}
      >
        {renderItem({
          item,
          // 【修复2】：传递真实的 isDragging 状态，而非硬编码 false
          isDragging,
          isOverlay: false,
          dragListeners: listeners || {},
          dragAttributes: attributes || {},
        })}
        {extraContent}
      </div>
    </div>
  )
}

// ==========================================
// VirtualSortableList 主组件
// ==========================================

export function VirtualSortableList<T extends { uuid: string }>({
  items,
  onItemsChange,
  renderItem,
  estimateSize = 72,
  overscan = 5,
  className = '',
  style,
  emptyContent,
  onDragStart,
  disableDrag = false,
  renderExtra,
  isItemExpanded,
}: VirtualSortableListProps<T>) {
  const [activeDragId, setActiveDragId] = useState<UniqueIdentifier | null>(null)
  const scrollRef = useRef<HTMLDivElement>(null)

  // 【修复6】：用 ref 跟踪最新的 items，避免 handleDragEnd 闭包陷阱
  const itemsRef = useRef(items)
  itemsRef.current = items

  // 【修复1核心】：拖拽开始时快照可见 items，拖拽结束前不再更新
  const frozenSortableItemsRef = useRef<string[] | null>(null)

  // 【修复5】：缓存拖拽中的 activeItem，用于 drop animation 期间保持 overlay 内容
  const activeItemCacheRef = useRef<T | null>(null)

  // DND 传感器配置
  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 5 } }),
    useSensor(KeyboardSensor, { coordinateGetter: sortableKeyboardCoordinates })
  )

  // 初始化虚拟列表
  const virtualizer = useVirtualizer({
    count: items.length,
    getScrollElement: () => scrollRef.current,
    estimateSize: () => estimateSize,
    overscan,
    getItemKey: (index) => items[index].uuid,
  })

  // 拖拽事件处理
  const handleDragStart = useCallback(
    (event: DragStartEvent) => {
      setActiveDragId(event.active.id)
      onDragStart?.()
    },
    [onDragStart]
  )

  // 【修复6】：使用 itemsRef.current 替代闭包中的 items
  const handleDragEnd = useCallback(
    (event: DragEndEvent) => {
      const { active, over } = event
      // 【修复5】：不在这里清除 activeDragId，让 DragOverlay 完成 drop animation
      // 先计算新顺序，再通过 setTimeout 清除拖拽状态
      if (over && active.id !== over.id) {
        const currentItems = itemsRef.current
        const oldIndex = currentItems.findIndex((g) => g.uuid === active.id)
        const newIndex = currentItems.findIndex((g) => g.uuid === over.id)
        if (oldIndex !== -1 && newIndex !== -1) {
          const newItems = arrayMove(currentItems, oldIndex, newIndex)
          onItemsChange(newItems)
        }
      }
      // 延迟清除拖拽状态，让 DragOverlay 完成 drop 动画
      frozenSortableItemsRef.current = null
      // 给 DragOverlay 畏够时间完成动画后清除
      setTimeout(() => {
        setActiveDragId(null)
        activeItemCacheRef.current = null
      }, 200)
    },
    [onItemsChange]
  )

  const handleDragCancel = useCallback(() => {
    frozenSortableItemsRef.current = null
    setActiveDragId(null)
    activeItemCacheRef.current = null
  }, [])

  // 【修复5】：计算 activeItem，带缓存
  const activeItem = useMemo(() => {
    if (!activeDragId) return activeItemCacheRef.current
    const found = items.find((g) => g.uuid === activeDragId) || null
    if (found) activeItemCacheRef.current = found
    return found
  }, [items, activeDragId])

  // 计算当前可见的虚拟节点
  const virtualItems = virtualizer.getVirtualItems()

  // 【修复1核心】：
  // - 非拖拽时：动态计算可见节点（正常行为）
  // - 拖拽时：使用快照值，滚动不会改变 SortableContext 的 items，避免状态破坏
  const computedSortableItems = useMemo(
    () => virtualItems.map((v) => items[v.index]?.uuid).filter(Boolean) as string[],
    [virtualItems, items]
  )

  const activeSortableItems = useMemo(() => {
    if (activeDragId) {
      // 拖拽进行中：如果还没有快照，创建一个；之后冻结不变
      if (!frozenSortableItemsRef.current) {
        frozenSortableItemsRef.current = computedSortableItems
      }
      return frozenSortableItemsRef.current
    }
    // 非拖拽状态：使用动态计算值
    frozenSortableItemsRef.current = null
    return computedSortableItems
  }, [activeDragId, computedSortableItems])

  // 空状态
  if (items.length === 0 && emptyContent) {
    return <div className={className} style={style}>{emptyContent}</div>
  }

  return (
    <div className={className} style={style}>
      <DndContext
        sensors={sensors}
        collisionDetection={closestCenter}
        onDragStart={handleDragStart}
        onDragEnd={handleDragEnd}
        onDragCancel={handleDragCancel}
      >
        {/* 滚动视窗 */}
        <div
          ref={scrollRef}
          className="flex-1 overflow-y-auto pr-2 relative custom-scrollbar"
          style={{ minHeight: 0 }}
        >
          {/* 占位层：撑开滚动条 */}
          <div style={{ height: virtualizer.getTotalSize(), width: '100%', position: 'relative' }}>
            <SortableContext items={activeSortableItems} strategy={verticalListSortingStrategy}>
              {virtualItems.map((virtualItem) => {
                const item = items[virtualItem.index]
                if (!item) return null

                return (
                  <SortableVirtualItem
                    key={item.uuid}
                    item={item}
                    virtualItem={virtualItem}
                    virtualizer={virtualizer}
                    renderItem={renderItem}
                    renderExtra={renderExtra}
                    disableDrag={disableDrag}
                    isExpanded={isItemExpanded?.(item) ?? false}
                  />
                )
              })}
            </SortableContext>
          </div>
        </div>

        {/* 悬浮层：拖拽时的替身 */}
        <DragOverlay
          dropAnimation={{
            sideEffects: defaultDropAnimationSideEffects({
              styles: { active: { opacity: '0.4' } },
            }),
          }}
        >
          {activeItem ? (
            <div style={{ cursor: 'grabbing' }}>
              {renderItem({
                item: activeItem,
                isDragging: true,
                isOverlay: true,
                dragListeners: {},
                dragAttributes: {},
              })}
            </div>
          ) : null}
        </DragOverlay>
      </DndContext>
    </div>
  )
}

// ==========================================
// 辅助组件：默认拖拽手柄
// ==========================================

export function DefaultDragHandle({
  listeners,
  attributes,
}: {
  listeners?: Record<string, any>
  attributes?: Record<string, any>
}) {
  return (
    <div
      {...attributes}
      {...listeners}
      className="cursor-grab active:cursor-grabbing p-1 rounded-md shrink-0 text-gray-400 hover:text-gray-700 hover:bg-gray-100"
    >
      <GripVertical size={14} />
    </div>
  )
}