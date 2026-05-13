import React, { useState, useCallback, useRef, useMemo, useLayoutEffect } from 'react'
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
  /** 外部传入的滚动容器 ref（用于从外部控制滚动位置） */
  scrollElementRef?: React.RefObject<HTMLDivElement | null>
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

  // 外层：文档流中、挂 measureElement ref 做自动测量（内部用 ResizeObserver）
  // 间距由 virtualizer 的 gap: 8 统一管理，无需在此设置 padding/margin
  return (
    <div
      ref={virtualizer.measureElement}
      data-index={virtualItem.index}
      style={{
        overflow: 'visible',
        position: 'relative',
        zIndex: isDragging ? 50 : isExpanded ? 40 : 1,
      }}
    >
      {/* 内层容器 - DND Layer: 负责拖拽时的物理变形和占位 */}
      <div
        ref={setNodeRef}
        style={{
          transform: CSS.Translate.toString(transform),
          transition,
          opacity: isDragging ? 0.3 : 1,
          position: 'relative',
          zIndex: isExpanded ? 1 : 'auto',
        }}
      >
        {renderItem({
          item,
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
  scrollElementRef,
}: VirtualSortableListProps<T>) {
  const [activeDragId, setActiveDragId] = useState<UniqueIdentifier | null>(null)
  const internalScrollRef = useRef<HTMLDivElement>(null)
  const scrollRef = scrollElementRef ?? internalScrollRef

  // 用 ref 跟踪最新的 items，避免 handleDragEnd 闭包陷阱
  const itemsRef = useRef(items)
  itemsRef.current = items

  // 缓存拖拽中的 activeItem，用于 drop animation 期间保持 overlay 内容
  const activeItemCacheRef = useRef<T | null>(null)

  // 清除拖拽状态的 setTimeout 句柄，用于防止竞态
  const clearTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  // DND 传感器配置
  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 5 } }),
    useSensor(KeyboardSensor, { coordinateGetter: sortableKeyboardCoordinates })
  )

  // 初始化虚拟列表
  // gap: 8 让虚拟器在 start/end 计算中自动包含项间距，无需手动设 padding/margin
  const virtualizer = useVirtualizer({
    count: items.length,
    getScrollElement: () => scrollRef.current,
    estimateSize: () => estimateSize,
    overscan,
    gap: 8,
    getItemKey: (index) => items[index].uuid,
  })

  // SortableContext 始终使用全量 UUID 列表，避免拖拽开始时切换集合导致闪烁和错位。
  const allUuids = useMemo(() => items.map((i) => i.uuid), [items])

  // 展开状态追踪：当展开/收起面板时触发 virtualizer.measure() 清除缓存
  const expandedStateKey = useMemo(
    () => items.map((i) => (isItemExpanded?.(i) ? '1' : '0')).join(''),
    [items, isItemExpanded]
  )
  useLayoutEffect(() => {
    // 展开/收起后清除测量缓存，确保动态高度（如 DeleteConfirmBanner 动画）被正确测量
    virtualizer.measure()
  }, [expandedStateKey, virtualizer])

  // 拖拽事件处理
  const handleDragStart = useCallback(
    (event: DragStartEvent) => {
      // 清除任何残留的延迟清除
      if (clearTimeoutRef.current) {
        clearTimeout(clearTimeoutRef.current)
        clearTimeoutRef.current = null
      }
      setActiveDragId(event.active.id)
      onDragStart?.()
    },
    [onDragStart]
  )

  const handleDragEnd = useCallback(
    (event: DragEndEvent) => {
      const { active, over } = event
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
      clearTimeoutRef.current = setTimeout(() => {
        setActiveDragId(null)
        activeItemCacheRef.current = null
        clearTimeoutRef.current = null
      }, 200)
    },
    [onItemsChange]
  )

  const handleDragCancel = useCallback(() => {
    if (clearTimeoutRef.current) {
      clearTimeout(clearTimeoutRef.current)
      clearTimeoutRef.current = null
    }
    setActiveDragId(null)
    activeItemCacheRef.current = null
  }, [])

  // 计算 activeItem，带缓存
  const activeItem = useMemo(() => {
    if (!activeDragId) return activeItemCacheRef.current
    const found = items.find((g) => g.uuid === activeDragId) || null
    if (found) activeItemCacheRef.current = found
    return found
  }, [items, activeDragId])

  const virtualItems = virtualizer.getVirtualItems()

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
        {/* 滚动视窗：contain: strict + overflow-anchor: none 防止浏览器滚动锚定干扰 */}
        <div
          ref={scrollRef}
          className="flex-1 overflow-y-auto pr-2 relative custom-scrollbar"
          style={{ minHeight: 0, contain: 'strict', overflowAnchor: 'none' }}
        >
          {/* 占位层：撑开滚动条 */}
          <div style={{ height: virtualizer.getTotalSize(), width: '100%', position: 'relative' }}>
            <SortableContext items={allUuids} strategy={verticalListSortingStrategy}>
              {/* 整块平移层：block translation，替代逐项绝对定位 */}
              <div
                style={{
                  position: 'absolute',
                  top: 0,
                  left: 0,
                  width: '100%',
                  transform: `translateY(${virtualItems[0]?.start ?? 0}px)`,
                }}
              >
                {virtualItems.map((virtualItem) => {
                  const item = items[virtualItem.index]
                  if (!item) return null

                  return (
                    <SortableVirtualItem
                      key={virtualItem.key}
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
              </div>
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