<template>
  <b-input-group
    :id="`${projectPrefix}stack`"
    :class="{'input-group--stack': true, 'is-collapsed': isCollapsed, 'is-modified': isModified, 'is-updating': isUpdating}"
    role="tabpanel"
  >
    <b-form-select
      class="select-stack"
      v-model="selectedStackId"
      :disabled="!hasUnusedStacks || isUpdating"
      :required="true"
      @change="isModified=true"
      v-show="!isCollapsed"
      :ref="`${projectPrefix}-select`"
      autofocus
    >
      <option value="" disabled>
        {{hasUnusedStacks ? $t('projects.fieldStackAdd') : $t('projects.fieldStackAllAdded')}}
      </option>
      <option
        v-for="(item,stackId) in unusedStacks"
        :key="stackId"
        :value="stackId"
      >
        {{item.stack.attributes.stackname + (item.isRemoved ? ' (removed)': '') + (item.isDefault? ' (default)': '')}}
      </option>
    </b-form-select>
    <b-input-group-append>
      <b-button
        variant="outline-info"
        :title="isUpdating ? $t('process.updating') : (isCollapsed ? $t('projects.fieldStackAddOne') : (isModified ? $t('projects.fieldStackAddSelected'): $t('projects.fieldStackUnmodified')))"
        v-b-tooltip.hover
        :disabled="isUpdating"
        :class="{'btn--submit': true, 'btn--add': isCollapsed}"
        @click.prevent="onAddProjectStack"
      >
        <font-awesome-icon
          v-if="isUpdating"
          key="status-icon"
          icon="circle-notch"
          spin
        />
        <font-awesome-icon
          v-else
          key="status-icon"
          :icon="['fa', (isCollapsed ? 'layer-group' : (isModified ? 'check' : 'times'))]"
        />
        <span>{{(isCollapsed && !isUpdating) ? '+' : ''}}</span>
      </b-button>
    </b-input-group-append>
  </b-input-group>

</template>

<script>
import { ProjectGetters, ProjectActions } from '../../_store/method-names'

export default {
  name: 'ProjectStackAdd',
  inject: [
    'project',
    'projectPrefix'
  ],
  props: {},
  data () {
    return {
      id: this.project.id,
      selectedStackId: '',
      isCollapsed: true,
      isModified: false,
      isUpdating: false
    }
  },
  computed: {

    unusedStacks () { return this.$store.getters[ProjectGetters.UNUSED_STACKS](this.project) },
    hasUnusedStacks () { return Object.entries(this.unusedStacks).length > 0 }

  },
  methods: {

    async maybeAddProjectStack (stackId) {
      if (!stackId) {
        return
      }

      try {
        this.isUpdating = true
        await this.$store.dispatch(
          ProjectActions.ADD_STACK,
          {
            project: this.project,
            stackId
          }
        )

        this.isUpdating = false
        this.isCollapsed = true
        this.selectedStackId = ''
        this.isModified = false

        this.$emit('maybe-hide-alert', this.$t('projects.fieldStackAddSome'))
        this.$emit('added-stack', stackId)
      } catch (e) {
        console.error(e.message)
      }
    },

    onAddProjectStack () {
      if (this.isCollapsed) {
        this.isCollapsed = false
        this.$nextTick(() => {
          this.$refs[`${this.projectPrefix}-select`].$el.focus()
        })
      } else {
        if (this.isModified) {
          this.maybeAddProjectStack(this.selectedStackId)
        } else {
          this.isCollapsed = true
        }
      }
    }
  }
}
</script>
<style scoped>
  .btn--add {
    position:relative;
  }

  .btn--add svg {
    position: relative;
    left: -2px;
    top: 2px;
  }
  .btn--add span {
    position: absolute;
    right: 6px;
    font-size: 17px;
    top: 0px;
  }

  .btn-outline-info {
    border-color: #ced4da;
  }

  .is-collapsed .btn-outline-info {
    border-color: transparent;
    border-top-left-radius: 0.25rem;
    border-bottom-left-radius: 0.25rem;
  }

</style>
