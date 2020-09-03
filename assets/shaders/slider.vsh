#version 330

in vec3 in_position;
in vec2 center;

uniform mat4 proj;
uniform float scale;
uniform mat4 distort;

out float distance_inv;
void main()
{
    gl_Position = distort * proj * (vec4(scale * in_position.xy, in_position.z, 1.0)+vec4(center, 0.0, 0.0));
    distance_inv = 1.0f - in_position.z;
}