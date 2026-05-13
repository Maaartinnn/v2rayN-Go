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
}

function SortableVirtualItem<T extends { uuid: string }>({
  item,
  virtualItem,
  virtualizer,
  renderItem,
  renderExtra,
  disableDrag = false,
}: SortableVirtualItemProps<T>) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({
    id: item.uuid,
    disabled: disableDrag,
  })

  const extraContent = renderExtra?.(item)

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
        zIndex: isDragging ? 50 : 1,
      }}
    >
      {/* 内层容器 - DND Layer: 负责拖拽时的物理变形和占位 */}
      <div
        ref={setNodeRef}
        style={{
          transform: CSS.Transform.toString(transform),
          transition,
          opacity: isDragging ? 0.4 : 1,
          marginBottom: '8px',
        }}
      >
        {renderItem({
          item,
          isDragging: false,
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
}: VirtualSortableListProps<T>) {
  const [activeDragId, setActiveDragId] = useState<UniqueIdentifier | null>(null)
  const scrollRef = useRef<HTMLDivElement>(null)

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

  const handleDragEnd = useCallback(
    (event: DragEndEvent) => {
      const { active, over } = event
      setActiveDragId(null)
      if (!over || active.id === over.id) return

      const oldIndex = items.findIndex((g) => g.uuid === active.id)
      const newIndex = items.findIndex((g) => g.uuid === over.id)
      const newItems = arrayMove(items, oldIndex, newIndex)
      onItemsChange(newItems)
    },
    [items, onItemsChange]
  )

  const activeItem = useMemo(
    () => items.find((g) => g.uuid === activeDragId),
    [items, activeDragId]
  )

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
        onDragCancel={() => setActiveDragId(null)}
      >
        {/* 滚动视窗 */}
        <div
          ref={scrollRef}
          className="flex-1 overflow-y-auto pr-2 relative custom-scrollbar"
          style={{ minHeight: 0 }}
        >
        {/* 占位层：撑开滚动条 */}
        <div style={{ height: virtualizer.getTotalSize(), width: '100%', position: 'relative' }}>
          <SortableContext items={items.map((g) => g.uuid)} strategy={verticalListSortingStrategy}>
            {virtualizer.getVirtualItems().map((virtualItem) => {
              const item = items[virtualItem.index]
              return (
                <SortableVirtualItem
                  key={item.uuid}
                  item={item}
                  virtualItem={virtualItem}
                  virtualizer={virtualizer}
                  renderItem={renderItem}
                  renderExtra={renderExtra}
                  disableDrag={disableDrag}
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